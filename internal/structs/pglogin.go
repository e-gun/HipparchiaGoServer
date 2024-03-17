package structs

type PostgresLogin struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DBName string
}
