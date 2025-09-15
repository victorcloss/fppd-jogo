package main

import (
	"math/rand"
	"time"
)

var (
	Portal    = Elemento{'◎', CorVerde, CorPadrao, false}
	Armadilha = Elemento{'¤', CorVermelho, CorPadrao, true}
)

type MsgPortal struct {
	X, Y int
}

type MsgArmadilha struct {
	X, Y  int
	Ativa bool
}

func iniciarInimigoPatrulha(jogo *Jogo, x, y int, done chan bool) {
	go func() {
		dx := 1
		for {
			select {
			case <-done:
				return
			default:
				if jogoPodeMoverPara(jogo, x+dx, y) {
					jogoMoverElemento(jogo, x, y, dx, 0)
					x += dx
				} else {
					dx = -dx
				}
				interfaceDesenharJogo(jogo)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

func iniciarPortal(jogo *Jogo, portalChan chan MsgPortal, done chan bool) {
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				px := rand.Intn(len(jogo.Mapa[0]))
				py := rand.Intn(len(jogo.Mapa))
				if jogoPodeMoverPara(jogo, px, py) {
					jogo.Mapa[py][px] = Portal
					interfaceDesenharJogo(jogo)
					select {
					case <-time.After(5 * time.Second):
						jogo.Mapa[py][px] = Vazio
					case msg := <-portalChan:
						if msg.X == px && msg.Y == py {
							jogo.StatusMsg = "Portal usado"
							jogo.Mapa[py][px] = Vazio
						}
					}
					interfaceDesenharJogo(jogo)
				}
				time.Sleep(7 * time.Second)
			}
		}
	}()
}

func iniciarArmadilha(jogo *Jogo, armadilhaChan chan MsgArmadilha, done chan bool) {
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-armadilhaChan:
				if msg.Ativa && jogoPodeMoverPara(jogo, msg.X, msg.Y) {
					jogo.Mapa[msg.Y][msg.X] = Armadilha
					interfaceDesenharJogo(jogo)
				} else {
					jogo.Mapa[msg.Y][msg.X] = Vazio
					interfaceDesenharJogo(jogo)
				}
			case <-time.After(10 * time.Second):
				jogo.StatusMsg = "Armadilha desarmada"
			}
		}
	}()
}
