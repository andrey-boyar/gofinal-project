package moduls

// Scheduler структура для хранения информации о задаче.
type Scheduler struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
type TaskId struct {
	Id int `json:"id"`
}

// Errors структура для ответа с ошибкой
type Errors struct {
	Errors string `json:"error"`
}

// структура конфигурации для базы данных
type Config struct {
	Port   string `json:"port"`
	DBFile string `json:"db_file"`
	// JWTSecret string `json:"jwt_secret"`
	// Password string `json:"password"`
	// TestEnv  string `json:"test_env"`
}

type SchedulerList struct {
	Tasks []Scheduler `json:"tasks"`
}
