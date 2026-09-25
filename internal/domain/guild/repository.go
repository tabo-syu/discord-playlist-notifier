package guild

type Repository interface {
	Exists(id DiscordID) (bool, error)
	Add(guild *Guild) error
	FindByDiscordID(id DiscordID) (*Guild, error)
	Delete(guild *Guild) error
}
