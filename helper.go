package main

import (
	"fmt"
	"sync"

	gc "github.com/dragonfax/goncurses"
	"github.com/fatih/color"
)

type GameManager struct {
	Games map[string]*Game
}

func NewGameManager() *GameManager {
	return &GameManager{
		Games: map[string]*Game{},
	}
}

// getGameWithAvailability returns a reference to a game with available spots for
// players. If one does not exist, nil is returned.
func (gm *GameManager) getGameWithAvailability() *Game {
	var g *Game

	for _, game := range gm.Games {
		spots := game.AvailableColors()
		if len(spots) > 0 {
			g = game
			break
		}
	}

	return g
}

func (gm *GameManager) SessionCount() int {
	sum := 0
	for _, game := range gm.Games {
		sum += game.SessionCount()
	}
	return sum
}

func (gm *GameManager) GameCount() int {
	return len(gm.Games)
}

const (
	gameWidth  = 78
	gameHeight = 22

	keyW = 'w'
	keyA = 'a'
	keyS = 's'
	keyD = 'd'

	keyH = 'h'
	keyJ = 'j'
	keyK = 'k'
	keyL = 'l'

	keyComma = ','
	keyO     = 'o'
	keyE     = 'e'

	keyCtrlC  = 3
	keyEscape = 27
)

var mutex = &sync.Mutex{}

func (gm *GameManager) HandleChannel(c *gc.Screen, wait bool) {
	g := gm.getGameWithAvailability()
	if g == nil {
		g = NewGame(gameWidth, gameHeight)
		gm.Games[g.Name] = g

		go g.Run()
	}

	session := NewSession(c, g.WorldWidth(), g.WorldHeight(),
		g.AvailableColors()[0])

	window, err := gc.NewWindowSP(c, 0, 0, 0, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	session.win = window

	g.AddSession(session)

	handleSession := func() {
		for {
			mutex.Lock()

			session.c.Set()

			r := window.GetChar()
			if r == 0 {
				fmt.Println("error reading key from window")
				mutex.Unlock()
				continue
			}

			switch r {
			case keyW, keyK, keyComma:
				session.Player.HandleUp()
			case keyA, keyH:
				session.Player.HandleLeft()
			case keyS, keyJ, keyO:
				session.Player.HandleDown()
			case keyD, keyL, keyE:
				session.Player.HandleRight()
			case keyCtrlC, keyEscape:
				if g.SessionCount() == 1 {
					delete(gm.Games, g.Name)
				}
			}

			mutex.Unlock()
		}
		g.RemoveSession(session)
	}

	if wait {
		handleSession()
	} else {
		go handleSession()
	}
}

type Session struct {
	c   *gc.Screen
	win *gc.Window

	Player *Player
}

func NewSession(c *gc.Screen, worldWidth, worldHeight int,
	color color.Attribute) *Session {

	s := Session{c: c}
	s.newGame(worldWidth, worldHeight, color)

	return &s
}

func (s *Session) newGame(worldWidth, worldHeight int, color color.Attribute) {
	s.Player = NewPlayer(s, worldWidth, worldHeight, color)
}

func (s *Session) StartOver(worldWidth, worldHeight int) {
	s.newGame(worldWidth, worldHeight, s.Player.Color)
}
