package discord

import (
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func getDiscordUser(s *discordgo.Session, guildId string, username string, discriminator string) (*discordgo.User, error) {
	members, err := s.GuildMembersSearch(guildId, username, 100)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.User.Username == username && member.User.Discriminator == discriminator {
			return member.User, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func parseUsernameAndDiscriminatorFromHandle(handle string) (string, string, error) {
	i := strings.Index(handle, "#")
	if i > -1 {
		if i >= len(handle)-1 {
			return "", "", fmt.Errorf("handle must be in format username#discriminator")
		}
		return handle[:i], handle[i+1:], nil
	}
	return handle, "0", nil
}

func SendMessageToUser(message string, username string, discriminator string) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		return err
	}
	err = discord.Open()
	if err != nil {
		return err
	}
	defer discord.Close()

	garageId := os.Getenv("GARAGE_GUILD_ID")
	user, err := getDiscordUser(discord, garageId, username, discriminator)
	if err != nil {
		return err
	}
	dmChannel, err := discord.UserChannelCreate(user.ID)
	if err != nil {
		return err
	}
	_, err = discord.ChannelMessageSend(dmChannel.ID, message)
	return err
}

// Handle must be in format 'username#discriminator' or just 'username' if there is no discriminator.
func SendAPIKeyToUser(apiKey string, handle string) error {
	username, discriminator, err := parseUsernameAndDiscriminatorFromHandle(handle)
	if err != nil {
		return err
	}
	return SendMessageToUser("Here is your API key for Dungeons and trolls: `"+apiKey+"`", username, discriminator)
}
