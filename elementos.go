package main

import (
	"math/rand"
	"time"
)

// Elementos visuais adicionais
var (
	Portal    = Elemento{'O', CorVerde, CorPadrao, false}       // Mudado para 'O' para compatibilidade
	Armadilha = Elemento{'X', CorVermelho, CorPadrao, true}     // Mudado para 'X'
	Fantasma  = Elemento{'G', CorCinzaEscuro, CorPadrao, false} // Mudado para 'G' (Ghost)
	Tesouro   = Elemento{'$', CorVerde, CorPadrao, false}       // Mudado para '$'
	Guardian  = Elemento{'@', CorVermelho, CorPadrao, true}     // Mudado para '@'
)

// Estruturas de mensagens para comunicação entre elementos
type MsgPortal struct {
	X, Y int
	Cmd  string // "usar", "fechar"
}

type MsgArmadilha struct {
	X, Y  int
	Ativa bool
}

type MsgFantasma struct {
	Cmd              string // "patrulhar", "perseguir", "ocultar"
	PlayerX, PlayerY int
}

type MsgTesouro struct {
	X, Y     int
	Aparecer bool
}

type MsgGuardian struct {
	Cmd              string // "dormir", "despertar", "atacar"
	PlayerX, PlayerY int
	Alerta           bool
}

// Canal para exclusão mútua do mapa (proteção contra condições de corrida)
var mapaMutex = make(chan bool, 1)

// Função para inicializar o mutex do mapa
func iniciarMutexMapa() {
	mapaMutex <- true // Inicializa como disponível
}

// Função para obter acesso exclusivo ao mapa
func obterAcessoMapa() {
	<-mapaMutex
}

// Função para liberar acesso ao mapa
func liberarAcessoMapa() {
	mapaMutex <- true
}

// Função auxiliar para verificar se posição é válida e segura
func posicaoValida(x, y int, jogo *Jogo) bool {
	return x >= 0 && x < len(jogo.Mapa[0]) && y >= 0 && y < len(jogo.Mapa) &&
		x < 79 && y < 29 // Limites seguros do terminal
}

// ELEMENTO 1: Inimigo Patrulha (melhorado com proteção)
func iniciarInimigoPatrulha(jogo *Jogo, x, y int, done chan bool) {
	go func() {
		dx := 1
		ticker := time.NewTicker(800 * time.Millisecond) // Mais lento para evitar flickering
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				obterAcessoMapa()

				// Verifica se posição atual é válida
				if !posicaoValida(x, y, jogo) {
					liberarAcessoMapa()
					continue
				}

				novoX := x + dx
				if posicaoValida(novoX, y, jogo) && jogoPodeMoverPara(jogo, novoX, y) {
					// Remove inimigo da posição atual
					if jogo.Mapa[y][x].simbolo == Inimigo.simbolo {
						jogo.Mapa[y][x] = Vazio
					}
					x = novoX
					jogo.Mapa[y][x] = Inimigo
				} else {
					dx = -dx // Muda direção
				}

				liberarAcessoMapa()

				// Renderiza com delay para evitar spam
				time.Sleep(50 * time.Millisecond)
				interfaceDesenharJogo(jogo)
			}
		}
	}()
}

// ELEMENTO 2: Portal com Timeout (protegido contra corrupção)
func iniciarPortal(jogo *Jogo, portalChan chan MsgPortal, done chan bool) {
	go func() {
		ticker := time.NewTicker(10 * time.Second) // Mais lento para melhor observação
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// Tenta criar portal em posição aleatória (com limites seguros)
				tentativas := 0
				for tentativas < 20 {
					px := 5 + rand.Intn(70) // Evita bordas
					py := 5 + rand.Intn(20) // Evita bordas

					obterAcessoMapa()
					if posicaoValida(px, py, jogo) && jogoPodeMoverPara(jogo, px, py) {
						jogo.Mapa[py][px] = Portal
						jogo.StatusMsg = "Portal apareceu!"
						liberarAcessoMapa()
						interfaceDesenharJogo(jogo)

						// Aguarda uso do portal ou timeout
						portalUsado := false
						select {
						case msg := <-portalChan:
							if msg.X == px && msg.Y == py && msg.Cmd == "usar" {
								obterAcessoMapa()
								jogo.StatusMsg = "Portal usado! Teletransporte!"
								jogo.Mapa[py][px] = Vazio

								// Teletransporta para posição segura
								for i := 0; i < 10; i++ {
									nx := 5 + rand.Intn(70)
									ny := 5 + rand.Intn(20)
									if posicaoValida(nx, ny, jogo) && jogoPodeMoverPara(jogo, nx, ny) {
										jogo.PosX, jogo.PosY = nx, ny
										break
									}
								}
								liberarAcessoMapa()
								portalUsado = true
							}
						case <-time.After(7 * time.Second):
							obterAcessoMapa()
							if jogo.Mapa[py][px].simbolo == Portal.simbolo {
								jogo.Mapa[py][px] = Vazio
								jogo.StatusMsg = "Portal fechou automaticamente"
							}
							liberarAcessoMapa()
						}

						if portalUsado {
							interfaceDesenharJogo(jogo)
						}
						break
					} else {
						liberarAcessoMapa()
					}
					tentativas++
				}
			}
		}
	}()
}

// ELEMENTO 3: Fantasma que Escuta Múltiplos Canais (simplificado)
func iniciarFantasma(jogo *Jogo, fantasmaChan chan MsgFantasma, done chan bool) {
	go func() {
		x, y := 15, 15
		visivel := true
		perseguindo := false
		ticker := time.NewTicker(1 * time.Second) // Mais lento
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case msg := <-fantasmaChan:
				switch msg.Cmd {
				case "perseguir":
					perseguindo = true
					visivel = true
				case "patrulhar":
					perseguindo = false
				case "ocultar":
					visivel = false
				}
			case <-ticker.C:
				obterAcessoMapa()

				// Remove fantasma da posição atual se visível
				if visivel && posicaoValida(x, y, jogo) && jogo.Mapa[y][x].simbolo == Fantasma.simbolo {
					jogo.Mapa[y][x] = Vazio
				}

				// Movimento mais simples
				novoX, novoY := x, y
				if perseguindo {
					// Move um passo em direção ao jogador
					if jogo.PosX > x {
						novoX = x + 1
					} else if jogo.PosX < x {
						novoX = x - 1
					} else if jogo.PosY > y {
						novoY = y + 1
					} else if jogo.PosY < y {
						novoY = y - 1
					}
				} else {
					// Movimento aleatório simples
					moves := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {0, 0}}
					move := moves[rand.Intn(len(moves))]
					novoX = x + move[0]
					novoY = y + move[1]
				}

				// Verifica se nova posição é válida
				if posicaoValida(novoX, novoY, jogo) && jogoPodeMoverPara(jogo, novoX, novoY) {
					x, y = novoX, novoY
				}

				// Coloca fantasma na posição se visível
				if visivel && posicaoValida(x, y, jogo) {
					jogo.Mapa[y][x] = Fantasma
				}

				liberarAcessoMapa()
				interfaceDesenharJogo(jogo)
			}
		}
	}()
}

// ELEMENTO 4: Sistema de Armadilhas (simplificado)
func iniciarArmadilha(jogo *Jogo, armadilhaChan chan MsgArmadilha, done chan bool) {
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-armadilhaChan:
				obterAcessoMapa()
				if posicaoValida(msg.X, msg.Y, jogo) {
					if msg.Ativa && jogoPodeMoverPara(jogo, msg.X, msg.Y) {
						jogo.Mapa[msg.Y][msg.X] = Armadilha
						jogo.StatusMsg = "Armadilha ativada!"
					} else if !msg.Ativa {
						if jogo.Mapa[msg.Y][msg.X].simbolo == Armadilha.simbolo {
							jogo.Mapa[msg.Y][msg.X] = Vazio
						}
						jogo.StatusMsg = "Armadilha desarmada"
					}
				}
				liberarAcessoMapa()
				interfaceDesenharJogo(jogo)

				// Auto-desativação simplificada
				if msg.Ativa {
					go func(x, y int) {
						time.Sleep(6 * time.Second)
						obterAcessoMapa()
						if posicaoValida(x, y, jogo) && jogo.Mapa[y][x].simbolo == Armadilha.simbolo {
							jogo.Mapa[y][x] = Vazio
							jogo.StatusMsg = "Armadilha expirou"
						}
						liberarAcessoMapa()
						interfaceDesenharJogo(jogo)
					}(msg.X, msg.Y)
				}
			}
		}
	}()
}

// ELEMENTO 5: Tesouro (simplificado)
func iniciarTesouro(jogo *Jogo, tesouroChan chan MsgTesouro, done chan bool) {
	go func() {
		for {
			select {
			case <-done:
				return
			case msg := <-tesouroChan:
				obterAcessoMapa()
				if posicaoValida(msg.X, msg.Y, jogo) {
					if msg.Aparecer && jogoPodeMoverPara(jogo, msg.X, msg.Y) {
						jogo.Mapa[msg.Y][msg.X] = Tesouro
						jogo.StatusMsg = "Tesouro apareceu!"
					} else if !msg.Aparecer {
						if jogo.Mapa[msg.Y][msg.X].simbolo == Tesouro.simbolo {
							jogo.Mapa[msg.Y][msg.X] = Vazio
						}
						jogo.StatusMsg = "Tesouro coletado!"
					}
				}
				liberarAcessoMapa()
				interfaceDesenharJogo(jogo)
			}
		}
	}()
}

// ELEMENTO 6: Guardião (simplificado)
func iniciarGuardian(jogo *Jogo, guardianChan chan MsgGuardian, done chan bool) {
	go func() {
		x, y := 25, 10
		dormindo := true
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// Coloca guardião no mapa
		obterAcessoMapa()
		if posicaoValida(x, y, jogo) {
			jogo.Mapa[y][x] = Guardian
		}
		liberarAcessoMapa()

		for {
			select {
			case <-done:
				return
			case msg := <-guardianChan:
				switch msg.Cmd {
				case "despertar":
					dormindo = false
					jogo.StatusMsg = "Guardião despertou!"
				case "dormir":
					dormindo = true
					jogo.StatusMsg = "Guardião adormeceu"
				}
			case <-ticker.C:
				if !dormindo && posicaoValida(x, y, jogo) {
					// Verifica proximidade do jogador
					distX := abs(jogo.PosX - x)
					distY := abs(jogo.PosY - y)

					if distX <= 3 && distY <= 3 {
						jogo.StatusMsg = "Guardião te detectou!"
					}
				}
				interfaceDesenharJogo(jogo)
			}
		}
	}()
}

// SISTEMA DE CONTROLE CENTRAL (simplificado)
func iniciarControleCentral(jogo *Jogo, fantasmaChan chan MsgFantasma, guardianChan chan MsgGuardian,
	tesouroChan chan MsgTesouro, armadilhaChan chan MsgArmadilha, done chan bool) {
	go func() {
		ticker := time.NewTicker(3 * time.Second) // Mais lento
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// Controle do fantasma baseado na posição do jogador
				if jogo.PosX > 30 {
					select {
					case fantasmaChan <- MsgFantasma{Cmd: "perseguir", PlayerX: jogo.PosX, PlayerY: jogo.PosY}:
					default:
					}
				} else {
					select {
					case fantasmaChan <- MsgFantasma{Cmd: "patrulhar"}:
					default:
					}
				}

				// Controle do guardião
				if jogo.PosX > 20 && jogo.PosY < 15 {
					select {
					case guardianChan <- MsgGuardian{Cmd: "despertar", PlayerX: jogo.PosX, PlayerY: jogo.PosY}:
					default:
					}
				}

				// Spawna elementos com menor frequência
				if rand.Intn(20) == 0 {
					tx := 5 + rand.Intn(70)
					ty := 5 + rand.Intn(20)
					select {
					case tesouroChan <- MsgTesouro{X: tx, Y: ty, Aparecer: true}:
					default:
					}
				}

				if rand.Intn(25) == 0 {
					ax := jogo.PosX + rand.Intn(5) - 2
					ay := jogo.PosY + rand.Intn(5) - 2
					if posicaoValida(ax, ay, jogo) {
						select {
						case armadilhaChan <- MsgArmadilha{X: ax, Y: ay, Ativa: true}:
						default:
						}
					}
				}
			}
		}
	}()
}

// Função auxiliar para valor absoluto
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
