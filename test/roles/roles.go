package roles

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"log"
	"math/rand"
	"strings"
	"time"

	amongus "github.com/allen-b1/go-amongus"
	aurefs "github.com/allen-b1/go-amongus/refs"
)

const (
	RoleCrewmate int = 0
	RoleImpostor int = 1
)

var (
	CrewmateColor color.Color
	ImpostorColor color.Color
)

type Role struct {
	Name   string
	Helper string
	Count  int
	Color  color.Color
}

var Roles = make([]*Role, 2)

func Register(r *Role) int {
	Roles = append(Roles, r)
	return len(Roles) - 1
}

// Objects
var (
	au         *amongus.AmongUs
	client     aurefs.AmongUsClientRef
	meetingHud aurefs.MeetingHudRef
	gameData   aurefs.GameDataRef
)

// Global state
var (
	prevGameState aurefs.GameStates
	prevVoteState aurefs.VoteStates
)

// Game-specific state
var (
	playerRoles       []int            = nil
	introObjectsFound map[uintptr]bool = nil
	introStartTime    time.Time
)

func Init(au_ *amongus.AmongUs) {
	au = au_
	client = aurefs.AmongUsClientInstance(au).Deref()

	CrewmateColor = aurefs.PaletteCrewmateBlue(au).Read()
	ImpostorColor = aurefs.PaletteImpostorRed(au).Read()
}

func Loop() {
	if client.AU == nil {
		panic("Loop called before Init")
	}

	gameState := client.GameState().Read()

	if gameState != prevGameState {
		if gameState != aurefs.GameStatesNotJoined {
			gameData = aurefs.GameDataInstance(au).Deref()
		} else {
			gameData = aurefs.GameDataRef{}
		}
	}

	if gameState == aurefs.GameStatesStarted {
		mHudPtr := aurefs.MeetingHudInstance(au)
		if !mHudPtr.Null() {
			meetingHud = mHudPtr.Deref()
		} else {
			meetingHud = aurefs.MeetingHudRef{}
		}
	} else {
		meetingHud = aurefs.MeetingHudRef{}
	}

	voteState := aurefs.VoteStates(255)
	if meetingHud.AU != nil {
		voteState = meetingHud.State().Read()
	}

	if gameState != prevGameState {
		if gameState == aurefs.GameStatesNotJoined || gameState == aurefs.GameStatesJoined {
			playerRoles = nil
			introObjectsFound = nil
			introStartTime = time.Time{}
			log.Println("reset")
		}
	}

	if (gameState != prevGameState || voteState != prevVoteState) && gameState == aurefs.GameStatesStarted && playerRoles != nil {
		updateNameColors()
	}

	if playerRoles == nil && gameState == aurefs.GameStatesStarted {
		if assignRoles() {
			introStart()
			updateNameColors()
		}
	}
	if playerRoles != nil && gameState == aurefs.GameStatesStarted {
		introRun()
	}

	prevGameState = gameState
	prevVoteState = voteState
}

// checks if impostors have been assigned; if so, assigns roles
func assignRoles() bool {
	players := gameData.AllPlayers().Deref()
	length := int(players.Len().Read())
	item := players.Items()

	playerRoles = make([]int, length)
	hasImpostor := false
	crewmates := make([]uint8, 0)
	for i := 0; i < length; i++ {
		player := aurefs.PlayerInfoPtrRef(item).Deref()
		if player.IsImpostor().Read() {
			hasImpostor = true
			playerRoles[player.PlayerId().Read()] = RoleImpostor
		} else {
			crewmates = append(crewmates, player.PlayerId().Read())
		}
		item = item.Ref(4)
	}

	if !hasImpostor {
		playerRoles = nil
		return false
	}

	r := rand.New(rand.NewSource(55))

	for roleId, role := range Roles {
		if roleId < 2 {
			continue
		}

		for i := 0; i < role.Count; i++ {
			idx := r.Intn(len(crewmates))
			playerID := crewmates[idx]
			crewmates = append(crewmates[:idx], crewmates[idx+1:]...)
			playerRoles[int(playerID)] = roleId
		}
	}

	log.Println("assigned roles!", playerRoles)
	return true
}

func PlayerRole(playerId uint8) int {
	if int(playerId) >= len(playerRoles) {
		return -1
	}
	return playerRoles[int(playerId)]
}

func isIntroCutscene(r amongus.Ref) bool {
	classRef, err := r.TryRef(0xc, 0, 0)
	return err == nil && classRef.Addr == amongus.PtrRef(aurefs.PlayerControlLocalPlayer(r.AU).Deref().NameText().Deref()).Deref().Addr &&
		r.Ref(0, 0).Addr == aurefs.SigIntroCutscene(r.AU)
}

func introStart() {
	playerID := aurefs.PlayerControlLocalPlayer(au).Deref().PlayerId().Read()
	roleID := playerRoles[playerID]
	if roleID < 2 {
		return
	}
	role := Roles[roleID]

	aurefs.PaletteCrewmateBlue(au).Write(role.Color)

	introStartTime = time.Now()
	introObjectsFound = make(map[uintptr]bool)
}

func introRun() {
	if introObjectsFound == nil {
		return
	}

	playerID := aurefs.PlayerControlLocalPlayer(au).Deref().PlayerId().Read()
	roleID := playerRoles[playerID]
	// assert roleID >= 2
	role := Roles[roleID]

	sigCutscene := aurefs.SigIntroCutscene(au)
	toSearch := new(bytes.Buffer)
	binary.Write(toSearch, binary.LittleEndian, uint32(sigCutscene))
	toSearch.Write([]byte{0, 0, 0, 0})
	ref, ok := au.Find(toSearch.Bytes(), 0x15000000, 0x35000000, func(ref amongus.Ref) bool {
		return isIntroCutscene(ref) && !introObjectsFound[ref.Addr]
	})
	if ok {
		cutref := aurefs.IntroCutsceneRef(ref)
		cutref.Title().Deref().Color().Write(role.Color)

		titleText := cutref.Title().Deref().Text().Deref()
		padding := (int(titleText.Len().Read()) - len(role.Name)) / 2
		if padding < 0 {
			padding = 0
		}
		titleText.Write(strings.Repeat(" ", padding) + role.Name)

		impostorText := cutref.ImpostorText().Deref().Text().Deref()
		padding = (int(impostorText.Len().Read()) - len(role.Helper)) / 2
		if padding < 0 {
			padding = 0
		}
		impostorText.Write(strings.Repeat(" ", padding) + role.Helper)

		introObjectsFound[ref.Addr] = true
	}

	if len(introObjectsFound) >= 3 || time.Now().Sub(introStartTime) >= 10*time.Second {
		log.Println("end cutscene")
		introObjectsFound = nil
		aurefs.PaletteCrewmateBlue(au).Write(CrewmateColor)
	}
}

func updateNameColors() {
	localPlayer := aurefs.PlayerControlLocalPlayer(au).Deref()
	playerID := localPlayer.PlayerId().Read()
	roleID := playerRoles[playerID]
	if roleID < 2 {
		return
	}
	role := Roles[roleID]

	localPlayer.NameText().Deref().Color().Write(role.Color)
	if meetingHud.AU != nil {
		playerStates := meetingHud.PlayerStates().Deref()
		items := playerStates.Items()
		length := int(playerStates.Len().Read())
		for i := 0; i < length; i++ {
			playerStatePtr := aurefs.PlayerVoteAreaPtrRef(items)
			if !playerStatePtr.Null() {
				playerState := playerStatePtr.Deref()
				if playerState.TargetPlayerId().Read() == int8(playerID) {
					playerState.NameText().Deref().Color().Write(role.Color)
					break
				}
			}
			items = items.Ref(4)
		}
	}
}
