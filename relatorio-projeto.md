# Relatório do Projeto: Elementos Autônomos Concorrentes para Jogo de Terminal

## Visão Geral

Este projeto implementa um sistema de elementos autônomos concorrentes para um jogo de terminal em Go, utilizando goroutines e canais para criar comportamentos dinâmicos e interativos. O sistema atende a todos os requisitos especificados, demonstrando conceitos fundamentais de programação paralela e distribuída.

## Elementos Implementados

### 1. Inimigo Patrulha
- **Comportamento**: Move-se horizontalmente de forma autônoma, mudando de direção ao encontrar obstáculos
- **Concorrência**: Executa em goroutine separada com ticker de 500ms
- **Sincronização**: Utiliza sistema de exclusão mútua via canais para acesso seguro ao mapa

### 2. Sistema de Portal com Timeout
- **Comportamento**: Aparece em posições aleatórias a cada 8 segundos, permanece ativo por 5 segundos
- **Comunicação**: Recebe mensagens via canal `MsgPortal` para ativação
- **Timeout**: Auto-fechamento após tempo limite se não for utilizado
- **Funcionalidade**: Teletransporta o jogador para posição aleatória quando usado

### 3. Fantasma (Escuta Múltiplos Canais)
- **Comportamento**: Alterna entre patrulha aleatória e perseguição do jogador
- **Múltiplos Canais**: Escuta comandos via `MsgFantasma` e timer interno simultaneamente
- **Estados**: Visível/invisível, patrulhar/perseguir
- **Select**: Utiliza `select` para reagir a diferentes eventos concorrentemente

### 4. Sistema de Armadilhas
- **Comportamento**: Ativadas via mensagens, desativam-se automaticamente após 8 segundos
- **Comunicação**: Controladas por canal `MsgArmadilha`
- **Timeout**: Auto-desativação implementada com goroutine temporizada

### 5. Tesouro Controlado por Mensagens
- **Comportamento**: Aparece e desaparece baseado em mensagens do sistema de controle
- **Comunicação**: Exclusivamente via canal `MsgTesouro`
- **Interação**: Coletável pelo jogador através da tecla de interação

### 6. Guardião com Detecção de Proximidade
- **Comportamento**: Detecta jogador em raio de 3 células, alterna entre dormindo/alerta
- **Múltiplos Canais**: Escuta comandos e monitora timer simultaneamente
- **Timeout**: Sistema de alerta com timeout de 2 segundos

## Recursos de Concorrência Implementados

### Exclusão Mútua
- **Implementação**: Canal `mapaMutex` com capacidade 1
- **Funções**: `obterAcessoMapa()` e `liberarAcessoMapa()`
- **Proteção**: Todas as modificações do mapa são protegidas contra condições de corrida
- **Conformidade**: Utiliza apenas canais, sem outros mecanismos de sincronização

### Comunicação Entre Elementos
- **Canais Implementados**:
  - `MsgPortal`: Comandos de uso do portal
  - `MsgArmadilha`: Ativação/desativação de armadilhas
  - `MsgFantasma`: Controle de estado do fantasma
  - `MsgTesouro`: Aparição/coleta de tesouros
  - `MsgGuardian`: Controle do guardião
- **Padrão**: Elementos comunicam-se via mensagens estruturadas ao invés de acesso direto

### Escuta Concorrente de Múltiplos Canais
- **Fantasma**: Escuta canal de comandos + timer de movimento
- **Guardião**: Escuta canal de comandos + timer de detecção + timeout de alerta
- **Portal**: Escuta canal de uso + timer de criação + timeout de fechamento
- **Implementação**: Uso extensivo de `select` com múltiplos `case` para gerenciar eventos simultâneos

### Comunicação com Timeout
- **Portal**: Timeout de 5 segundos para fechamento automático
- **Armadilha**: Timeout de 8 segundos para desativação automática
- **Guardião**: Timeout de 2 segundos para perder alerta
- **Implementação**: `time.After()` em combinação com `select`

## Sistema de Controle Central

### Coordenação de Elementos
- **Função**: `iniciarControleCentral()`
- **Responsabilidade**: Monitora estado do jogo e envia comandos para elementos
- **Lógica**: 
  - Fantasma persegue quando jogador está na área direita (x > 30)
  - Guardião desperta quando jogador se aproxima (x > 20, y < 15)
  - Spawna tesouros aleatoriamente
  - Ativa armadilhas próximas ao jogador

### Gerenciamento de Interações
- **Função**: `gerenciarInteracoes()`
- **Auto-interações**: Portal e tesouro são automaticamente utilizados quando jogador os toca
- **Detecção**: Monitora posição do jogador continuamente

## Estruturas de Dados

### Mensagens
```go
type MsgPortal struct {
    X, Y int
    Cmd  string // "usar", "fechar"
}

type MsgFantasma struct {
    Cmd           string // "patrulhar", "perseguir", "ocultar"
    PlayerX, PlayerY int
}
```

### Elementos Visuais
- **Portal**: ◎ (verde) - Teletransporte
- **Armadilha**: ¤ (vermelho) - Perigo temporário
- **Fantasma**: ♠ (cinza) - Entidade móvel
- **Tesouro**: ♦ (verde) - Coletável
- **Guardião**: ☢ (vermelho) - Sentinela

## Características Técnicas

### Sincronização Robusta
- **Deadlock Prevention**: Uso de `defer` para garantir liberação de recursos
- **Non-blocking Channels**: `select` com `default` para evitar bloqueios
- **Graceful Shutdown**: Canal `done` para encerramento coordenado

### Performance
- **Timers Otimizados**: Diferentes frequências para diferentes elementos
- **Buffer de Canais**: Canais com buffer para reduzir bloqueios
- **Gestão de Recursos**: Cleanup adequado de goroutines e timers

### Robustez
- **Verificação de Limites**: Todas as operações de movimento verificam limites do mapa
- **Fallback**: Comportamentos alternativos quando canais estão cheios
- **Error Handling**: Tratamento de situações excepcionais

## Atendimento aos Requisitos

### ✅ Requisitos Atendidos

1. **Mínimo 3 tipos de elementos concorrentes**: ✅ 6 tipos implementados
2. **Execução em goroutines separadas**: ✅ Cada elemento em sua própria goroutine
3. **Comportamentos visíveis e distintos**: ✅ Cada elemento tem comportamento único
4. **Exclusão mútua usando canais**: ✅ Sistema `mapaMutex` implementado
5. **Comunicação entre elementos por canais**: ✅ Sistema completo de mensagens
6. **Escuta de múltiplos canais**: ✅ Fantasma, Guardião e Portal implementados
7. **Comunicação com timeout**: ✅ Portal, Armadilha e Guardião implementados

### Funcionalidades Extras

- **Sistema de Controle Central**: Coordena todos os elementos automaticamente
- **Interações Automáticas**: Elementos reagem à presença do jogador
- **Estados Complexos**: Elementos com múltiplos estados e transições
- **Efeitos Visuais**: Mensagens de status dinâmicas e informativas

## Compilação e Execução

### Pré-requisitos
```bash
go mod init jogo
go get -u github.com/nsf/termbox-go
```

### Compilação
```bash
go build -o jogo
```

### Execução
```bash
./jogo [arquivo_mapa]
```

### Controles
- **W/A/S/D**: Movimento
- **E**: Interação
- **ESC**: Sair

## Conclusão

O projeto implementa com sucesso um sistema complexo de elementos concorrentes que demonstra os principais conceitos de programação paralela em Go. A utilização exclusiva de canais para sincronização e comunicação resulta em um código idiomático e eficiente, seguindo as melhores práticas da linguagem Go.

O sistema é extensível e permite fácil adição de novos elementos, mantendo a arquitetura de comunicação por mensagens e o padrão de exclusão mútua estabelecido.