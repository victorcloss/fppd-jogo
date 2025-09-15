// interface.go - Interface gráfica do jogo usando termbox com proteção contra corrupção
package main

import (
	"github.com/nsf/termbox-go"
)

// Define um tipo Cor para encapsuladar as cores do termbox
type Cor = termbox.Attribute

// Definições de cores utilizadas no jogo
const (
	CorPadrao      Cor = termbox.ColorDefault
	CorCinzaEscuro     = termbox.ColorDarkGray
	CorVermelho        = termbox.ColorRed
	CorVerde           = termbox.ColorGreen
	CorParede          = termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim
	CorFundoParede     = termbox.ColorDarkGray
	CorTexto           = termbox.ColorDarkGray
)

// EventoTeclado representa uma ação detectada do teclado
type EventoTeclado struct {
	Tipo  string // "sair", "interagir", "mover"
	Tecla rune   // Tecla pressionada, usada no caso de movimento
}

// Canal para serializar operações de desenho (evita corrupção visual)
var canalDesenho = make(chan func(), 100)
var desenhoAtivo = false

// Inicializa a interface gráfica usando termbox
func interfaceIniciar() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	// Inicia o worker de desenho em goroutine separada
	iniciarWorkerDesenho()
}

// Worker que processa todas as operações de desenho sequencialmente
func iniciarWorkerDesenho() {
	desenhoAtivo = true
	go func() {
		for operacao := range canalDesenho {
			if operacao != nil {
				operacao()
			}
		}
		desenhoAtivo = false
	}()
}

// Encerra o uso da interface termbox
func interfaceFinalizar() {
	// Para o worker de desenho
	if desenhoAtivo {
		close(canalDesenho)
		// Aguarda worker finalizar
		for desenhoAtivo {
			// Busy wait simples
		}
	}
	termbox.Close()
}

// Lê um evento do teclado e o traduz para um EventoTeclado
func interfaceLerEventoTeclado() EventoTeclado {
	ev := termbox.PollEvent()
	if ev.Type != termbox.EventKey {
		return EventoTeclado{}
	}
	if ev.Key == termbox.KeyEsc {
		return EventoTeclado{Tipo: "sair"}
	}
	if ev.Ch == 'e' || ev.Ch == 'E' {
		return EventoTeclado{Tipo: "interagir"}
	}
	return EventoTeclado{Tipo: "mover", Tecla: ev.Ch}
}

// Renderiza todo o estado atual do jogo na tela de forma thread-safe
func interfaceDesenharJogo(jogo *Jogo) {
	// Envia operação de desenho para o worker
	select {
	case canalDesenho <- func() { renderizarJogoSeguro(jogo) }:
	default:
		// Se canal estiver cheio, ignora esta renderização para evitar bloqueio
	}
}

// Função interna que faz a renderização real (executada pelo worker)
func renderizarJogoSeguro(jogo *Jogo) {
	// Obtem acesso exclusivo ao estado do jogo
	obterAcessoMapa()
	defer liberarAcessoMapa()

	// Cria uma cópia local do estado para renderização
	mapaLocal := make([][]Elemento, len(jogo.Mapa))
	for i := range jogo.Mapa {
		mapaLocal[i] = make([]Elemento, len(jogo.Mapa[i]))
		copy(mapaLocal[i], jogo.Mapa[i])
	}
	posX, posY := jogo.PosX, jogo.PosY
	statusMsg := jogo.StatusMsg

	// Limpa a tela
	termbox.Clear(CorPadrao, CorPadrao)

	// Desenha todos os elementos do mapa
	for y, linha := range mapaLocal {
		for x, elem := range linha {
			if x < 80 && y < 30 { // Limita às dimensões seguras do terminal
				termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
			}
		}
	}

	// Desenha o personagem sobre o mapa (se estiver em posição válida)
	if posX >= 0 && posX < 80 && posY >= 0 && posY < 30 {
		termbox.SetCell(posX, posY, Personagem.simbolo, Personagem.cor, Personagem.corFundo)
	}

	// Desenha a barra de status
	desenharBarraDeStatusSegura(statusMsg, len(mapaLocal))

	// Força a atualização do terminal
	termbox.Flush()
}

// Exibe uma barra de status com informações úteis ao jogador
func desenharBarraDeStatusSegura(statusMsg string, alturaJogo int) {
	// Limita o tamanho da mensagem para evitar overflow
	if len(statusMsg) > 78 {
		statusMsg = statusMsg[:78]
	}

	// Linha de status dinâmica
	linhaStatus := alturaJogo + 1
	if linhaStatus < 30 { // Verifica se cabe na tela
		for i, c := range statusMsg {
			if i < 79 { // Limita largura
				termbox.SetCell(i, linhaStatus, c, CorTexto, CorPadrao)
			}
		}
	}

	// Instruções fixas
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	linhaInstrucoes := alturaJogo + 3
	if linhaInstrucoes < 30 && len(msg) < 79 {
		for i, c := range msg {
			termbox.SetCell(i, linhaInstrucoes, c, CorTexto, CorPadrao)
		}
	}
}

// Limpa a tela do terminal (função de conveniência)
func interfaceLimparTela() {
	select {
	case canalDesenho <- func() { termbox.Clear(CorPadrao, CorPadrao) }:
	default:
	}
}

// Força a atualização da tela do terminal (função de conveniência)
func interfaceAtualizarTela() {
	select {
	case canalDesenho <- func() { termbox.Flush() }:
	default:
	}
}

// Desenha um elemento na posição (x, y) de forma segura
func interfaceDesenharElemento(x, y int, elem Elemento) {
	select {
	case canalDesenho <- func() {
		if x >= 0 && x < 80 && y >= 0 && y < 30 {
			termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
		}
	}:
	default:
	}
}
