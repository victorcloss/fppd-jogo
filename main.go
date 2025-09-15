// main.go - Loop principal do jogo com elementos concorrentes completos
package main

import (
	"os"
	"time"
)

func main() {
	interfaceIniciar()
	defer interfaceFinalizar()

	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Inicializa o sistema de exclusão mútua
	iniciarMutexMapa()

	// Criação dos canais para comunicação entre elementos
	portalChan := make(chan MsgPortal, 5)
	armadilhaChan := make(chan MsgArmadilha, 10)
	fantasmaChan := make(chan MsgFantasma, 5)
	tesouroChan := make(chan MsgTesouro, 5)
	guardianChan := make(chan MsgGuardian, 5)
	done := make(chan bool)

	// Inicia todos os elementos concorrentes
	iniciarInimigoPatrulha(&jogo, 10, 5, done)
	iniciarPortal(&jogo, portalChan, done)
	iniciarArmadilha(&jogo, armadilhaChan, done)
	iniciarFantasma(&jogo, fantasmaChan, done)
	iniciarTesouro(&jogo, tesouroChan, done)
	iniciarGuardian(&jogo, guardianChan, done)

	// Inicia o sistema de controle central que coordena os elementos
	iniciarControleCentral(&jogo, fantasmaChan, guardianChan, tesouroChan, armadilhaChan, done)

	// Goroutine para gerenciar interações automáticas
	go gerenciarInteracoes(&jogo, portalChan, tesouroChan, done)

	// Primeira renderização
	interfaceDesenharJogo(&jogo)

	// Loop principal do jogo
	for {
		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, &jogo, portalChan, tesouroChan); !continuar {
			// Sinaliza para todas as goroutines pararem
			close(done)
			// Aguarda um pouco para as goroutines terminarem graciosamente
			time.Sleep(100 * time.Millisecond)
			break
		}
		interfaceDesenharJogo(&jogo)
	}
}

// Função para gerenciar interações automáticas baseadas na posição do jogador
func gerenciarInteracoes(jogo *Jogo, portalChan chan MsgPortal, tesouroChan chan MsgTesouro, done chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Verifica se jogador está sobre um portal
			obterAcessoMapa()
			elementoAtual := jogo.Mapa[jogo.PosY][jogo.PosX]

			if elementoAtual.simbolo == Portal.simbolo {
				// Auto-uso do portal após 1 segundo
				go func(x, y int) {
					time.Sleep(1 * time.Second)
					select {
					case portalChan <- MsgPortal{X: x, Y: y, Cmd: "usar"}:
					default:
					}
				}(jogo.PosX, jogo.PosY)
			}

			// Verifica se jogador está sobre um tesouro
			if elementoAtual.simbolo == Tesouro.simbolo {
				select {
				case tesouroChan <- MsgTesouro{X: jogo.PosX, Y: jogo.PosY, Aparecer: false}:
				default:
				}
			}

			liberarAcessoMapa()
		}
	}
}
