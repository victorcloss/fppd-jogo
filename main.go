// main.go - Loop principal do jogo
package main

import "os"

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

	portalChan := make(chan MsgPortal)
	armadilhaChan := make(chan MsgArmadilha)
	done := make(chan bool)

	iniciarInimigoPatrulha(&jogo, 10, 5, done)
	iniciarPortal(&jogo, portalChan, done)
	iniciarArmadilha(&jogo, armadilhaChan, done)

	interfaceDesenharJogo(&jogo)

	for {
		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, &jogo); !continuar {
			break
		}
		interfaceDesenharJogo(&jogo)
	}
}
