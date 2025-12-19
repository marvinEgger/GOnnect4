package lib

// MessageType identifies the type of websocket message
type MessageType string

const (
	// Client to Server
	MsgLogin      MessageType = "login"
	MsgCreateGame MessageType = "create_game"
	MsgJoinGame   MessageType = "join_game"
	MsgPlay       MessageType = "play"
	MsgReplay     MessageType = "replay"
	MsgForfeit    MessageType = "forfeit"
	MsgLeaveLobby MessageType = "leave_lobby"
	MsgDisconnect MessageType = "disconnect"

	// Server to Client
	MsgWelcome     MessageType = "welcome"
	MsgGameCreated MessageType = "game_created"
	MsgGameJoined  MessageType = "game_joined"
	MsgGameStart   MessageType = "game_start"
	MsgGameState   MessageType = "game_state"
	MsgMove        MessageType = "move"
	MsgGameOver    MessageType = "game_over"
	MsgReplayReq   MessageType = "replay_request"
	MsgError       MessageType = "error"
)

// Message represents a websocket message
type Message struct {
	Type MessageType `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// LoginData contains login credentials
type LoginData struct {
	Username string    `json:"username"`
	PlayerID *PlayerID `json:"player_id,omitempty"` // for reconnection
}

// WelcomeData sent after successful login
type WelcomeData struct {
	PlayerID PlayerID `json:"player_id"`
	Username string   `json:"username"`
}

// GameCreatedData sent when game is created
type GameCreatedData struct {
	Code string `json:"code"`
}

// JoinGameData contains game join request
type JoinGameData struct {
	Code string `json:"code"`
}

// GameJoinedData sent when successfully joined a game
type GameJoinedData struct {
	Code      string        `json:"code"`
	PlayerIdx int           `json:"player_idx"`
	Status    GameStatus    `json:"status"`
	Players   [2]PlayerInfo `json:"players"`
}

// PlayerInfo contains public player information
type PlayerInfo struct {
	ID        PlayerID `json:"id"`
	Username  string   `json:"username"`
	Connected bool     `json:"connected"`
}

// GameStartData sent when game starts
type GameStartData struct {
	Code          string        `json:"code"`
	CurrentTurn   int           `json:"current_turn"`
	Players       [2]PlayerInfo `json:"players"`
	TimeRemaining [2]int64      `json:"time_remaining"` // milliseconds
}

// PlayData contains a move request
type PlayData struct {
	Column int `json:"column"`
}

// MoveData broadcasts a move to both players
type MoveData struct {
	PlayerIdx     int              `json:"player_idx"`
	Column        int              `json:"column"`
	Row           int              `json:"row"`
	Board         [Rows][Cols]Cell `json:"board"`
	NextTurn      int              `json:"next_turn"`
	TimeRemaining [2]int64         `json:"time_remaining"` // milliseconds
}

// GameOverData sent when game ends
type GameOverData struct {
	Result GameResult       `json:"result"`
	Board  [Rows][Cols]Cell `json:"board"`
}

// ReplayRequestData sent when a player requests replay
type ReplayRequestData struct {
	PlayerIdx int `json:"player_idx"`
}

// ErrorData contains error information
type ErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// GameStateData contains full game state for reconnection
type GameStateData struct {
	Code           string           `json:"code"`
	Status         GameStatus       `json:"status"`
	Result         GameResult       `json:"result"`
	Board          [Rows][Cols]Cell `json:"board"`
	Players        [2]PlayerInfo    `json:"players"`
	PlayerIdx      int              `json:"player_idx"`
	CurrentTurn    int              `json:"current_turn"`
	MoveCount      int              `json:"move_count"`
	TimeRemaining  [2]int64         `json:"time_remaining"` // milliseconds
	ReplayRequests [2]bool          `json:"replay_requests"`
}
