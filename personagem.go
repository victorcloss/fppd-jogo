// personagem.go - Funções para movimentação e ações do personagem com interações expandidas
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1 // Move para cima
	case 'a':
		dx = -1 // Move para a esquerda
	case 's':
		dy = 1 // Move para baixo
	case 'd':
		dx = 1 // Move para a direita
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy

	// Usa exclusão mútua para proteger o acesso ao mapa
	obterAcessoMapa()
	defer liberarAcessoMapa()

	// Verifica se o movimento é permitido e realiza a movimentação
	if jogoPodeMoverPara(jogo, nx, ny) {
		// Verifica interações especiais antes de mover
		elementoDestino := jogo.Mapa[ny][nx]

		switch elementoDestino.simbolo {
		case Armadilha.simbolo:
			jogo.StatusMsg = "Você pisou numa armadilha! Cuidado!"
		case Fantasma.simbolo:
			jogo.StatusMsg = "Você passou através do fantasma... arrepiante!"
		}

		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny
	} else {
		// Verifica o que está bloqueando o movimento
		if nx >= 0 && nx < len(jogo.Mapa[0]) && ny >= 0 && ny < len(jogo.Mapa) {
			elementoBloqueador := jogo.Mapa[ny][nx]
			switch elementoBloqueador.simbolo {
			case Parede.simbolo:
				jogo.StatusMsg = "Você bateu na parede!"
			case Inimigo.simbolo:
				jogo.StatusMsg = "Um inimigo está bloqueando o caminho!"
			case Guardian.simbolo:
				jogo.StatusMsg = "O guardião não deixa você passar!"
			default:
				jogo.StatusMsg = "Caminho bloqueado!"
			}
		}
	}
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
func personagemInteragir(jogo *Jogo, portalChan chan MsgPortal, tesouroChan chan MsgTesouro) {
	obterAcessoMapa()
	elementoAtual := jogo.Mapa[jogo.PosY][jogo.PosX]
	liberarAcessoMapa()

	// Verifica interações baseadas no elemento atual
	switch elementoAtual.simbolo {
	case Portal.simbolo:
		jogo.StatusMsg = "Usando portal..."
		select {
		case portalChan <- MsgPortal{X: jogo.PosX, Y: jogo.PosY, Cmd: "usar"}:
		default:
			jogo.StatusMsg = "Portal não responde..."
		}
	case Tesouro.simbolo:
		jogo.StatusMsg = "Coletando tesouro!"
		select {
		case tesouroChan <- MsgTesouro{X: jogo.PosX, Y: jogo.PosY, Aparecer: false}:
		default:
			jogo.StatusMsg = "Não foi possível coletar o tesouro"
		}
	default:
		// Verifica elementos adjacentes para interação
		interagiu := false

		// Verifica as 4 direções adjacentes
		direccoes := [][]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
		for _, dir := range direccoes {
			x, y := jogo.PosX+dir[0], jogo.PosY+dir[1]
			if x >= 0 && x < len(jogo.Mapa[0]) && y >= 0 && y < len(jogo.Mapa) {
				obterAcessoMapa()
				elemento := jogo.Mapa[y][x]
				liberarAcessoMapa()

				switch elemento.simbolo {
				case Vegetacao.simbolo:
					if !interagiu {
						jogo.StatusMsg = "Você examina a vegetação... nada interessante"
						interagiu = true
					}
				case Parede.simbolo:
					if !interagiu {
						jogo.StatusMsg = "Você toca na parede... é sólida"
						interagiu = true
					}
				case Inimigo.simbolo:
					if !interagiu {
						jogo.StatusMsg = "O inimigo te olha ameaçadoramente!"
						interagiu = true
					}
				case Guardian.simbolo:
					if !interagiu {
						jogo.StatusMsg = "O guardião permanece imóvel... por enquanto"
						interagiu = true
					}
				}
			}
		}

		if !interagiu {
			jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d) - nada acontece", jogo.PosX, jogo.PosY)
		}
	}
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo, portalChan chan MsgPortal, tesouroChan chan MsgTesouro) bool {
	switch ev.Tipo {
	case "sair":
		jogo.StatusMsg = "Saindo do jogo..."
		return false
	case "interagir":
		personagemInteragir(jogo, portalChan, tesouroChan)
	case "mover":
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
