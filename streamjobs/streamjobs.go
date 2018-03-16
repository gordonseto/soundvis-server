package streamjobs

import (
	"github.com/gordonseto/soundvis-server/users/repositories"
	//"log"
	"fmt"
	"log"
	"github.com/gordonseto/soundvis-server/users/models"
)

type StreamJobManager struct {
	userRepository *usersrepository.UsersRepository
}

func NewStreamJobManager(ur *usersrepository.UsersRepository) *StreamJobManager {
	return &StreamJobManager{ur}
}

func StreamJobName() string {
	return "stream_job"
}

// fetches currentPlaying for all users with currentPlaying = true
// sends socket message and android notification if song has changed
func (sjm *StreamJobManager) RefreshNowPlaying() {
	users, err := sjm.userRepository.FindUsersByIsPlaying(true)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("Fetched new batch of users:")
	for _, user := range users {
		fmt.Println(user.Id)
		sjm.GetNowPlayingForUser(user)
	}
}

func (sjm *StreamJobManager) GetNowPlayingForUser(user models.User) {
}
