package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Score struct {
	FirstRound  int
	SecondRound int
	ThirdRound  int
	FourthRound int
}

type Team struct {
	Name string
	Score
	IsEliminated bool
	Played       bool
}

type QuestionOpts struct {
	PointPerAnswer  int
	TimePerQuestion *time.Duration // it should only be minutes
}

type Questions struct {
	QuestionOpts
	ListOfQuestions []Question
	File            *os.File
}

type Question struct {
	Question      string
	Answers       map[int]string
	CorrectAnswer int
}

func (q *Questions) ParseTrivia() {
	lineScanner := bufio.NewScanner(q.File)
	defer q.File.Close()

	for lineScanner.Scan() {
		line := lineScanner.Text()

		question := strings.Split(line, "=>")
		questionText := strings.Trim(question[0], " ")

		question[1] = strings.ReplaceAll(question[1], "[", "")
		question[1] = strings.ReplaceAll(question[1], "]", "")

		answers := strings.Split(question[1], ",")
		correctAnswer, _ := strconv.Atoi(strings.Trim(answers[len(answers)-1], " "))

		answers = answers[:len(answers)-1]

		answersMap := make(map[int]string, len(answers))
		for i, answer := range answers {
			answersMap[i] = strings.Trim(answer, " ")
		}

		questionObj := Question{
			Question:      questionText,
			Answers:       answersMap,
			CorrectAnswer: correctAnswer,
		}

		q.ListOfQuestions = append(q.ListOfQuestions, questionObj)
	}
}

func (q *Questions) GetQuestion(deleteLine bool) Question {
	rand.NewSource(time.Now().UnixNano())
	randNum := rand.Intn(len(q.ListOfQuestions))
	question := q.ListOfQuestions[randNum]

	if deleteLine {
		questionFile, _ := os.OpenFile("./Teams/trivia.md", os.O_RDWR, 0644)
		scanner := bufio.NewScanner(questionFile)

		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, q.ListOfQuestions[randNum].Question) || line == "" {
				continue
			}

			lines = append(lines, line)
		}

		out, err := os.OpenFile("./Teams/trivia.md", os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal("Error Opening Out file: ", err)
		}
		defer out.Close()

		for _, line := range lines {
			if _, err = out.WriteString(line + "\n"); err != nil {
				log.Fatal("Error Writing to Out file: ", err)
			}
		}

		q.ListOfQuestions = slices.Delete(q.ListOfQuestions, randNum, randNum+1)
	}

	return question
}

func (t Team) Update() {
	// update the team's line to the current state
	// if the team has played set played to true

	file, err := os.OpenFile("./Teams/teams.md", os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewScanner(file)
	var lines []string
	for reader.Scan() {
		line := reader.Text()
		if strings.Contains(line, t.Name) {
			line = t.Name + "\t\t" + "==> [" + strconv.Itoa(t.FirstRound) + ", " + strconv.Itoa(t.SecondRound) + ", " + strconv.Itoa(t.ThirdRound) + ", " + strconv.Itoa(t.FourthRound) + "] | " + strconv.FormatBool(t.IsEliminated) + " | " + strconv.FormatBool(t.Played)
		}

		lines = append(lines, line)
	}

	out, err := os.OpenFile("./Teams/teams.md", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("Error Opening Out file: ", err)
	}
	defer out.Close()

	for _, line := range lines {
		if _, err = out.WriteString(line + "\n"); err != nil {
			log.Fatal("Error Writing to Out file: ", err)
		}
	}
}

func teamParser() ([]Team, error) {
	teamFile, err := os.Open("./Teams/teams.md")
	if err != nil {
		return nil, err
	}

	reader := bufio.NewScanner(teamFile)
	// skip the next 3 lines because they are the headers
	for i := 0; i < 3; i++ {
		reader.Scan()
	}

	var teams []Team
	for reader.Scan() {
		line := reader.Text()

		if line == "" {
			continue
		}

		team := strings.Split(line, "==>")
		teamScore := strings.Split(team[1], "|")
		teamScore[0] = strings.ReplaceAll(teamScore[0], "[", "")
		teamScore[0] = strings.ReplaceAll(teamScore[0], "]", "")
		scores := strings.Split(teamScore[0], ",")

		eliminated := strings.Trim(teamScore[1], " ")
		eliminatedBool, _ := strconv.ParseBool(eliminated)

		played := strings.Trim(teamScore[2], " ")
		playedBool, _ := strconv.ParseBool(played)

		teamScoreInt := make([]int, 4)
		for i, score := range scores {
			score = strings.TrimSpace(score)
			teamScoreInt[i], err = strconv.Atoi(score)
			if err != nil {
				return nil, err
			} 
		}

		teamObj := Team{
			Name: strings.TrimSpace(team[0]),
			Score: Score{
				FirstRound:  teamScoreInt[0],
				SecondRound: teamScoreInt[1],
				ThirdRound:  teamScoreInt[2],
				FourthRound: teamScoreInt[3],
			},
			IsEliminated: eliminatedBool,
			Played:       playedBool,
		}

		teams = append(teams, teamObj)
	}

	return teams, nil
}

func addTeam(team *Team) error {
	teamFile, err := os.OpenFile("./Teams/teams.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	parsedTeams, err := teamParser()
	if err != nil {
		return err
	}

	for _, t := range parsedTeams {
		if t.Name == team.Name {
			return fmt.Errorf("Team already exists")
		}
	}

	_, err = teamFile.WriteString(team.Name + "\t\t" + "==> [0, 0, 0, 0] | false | false\n")
	if err != nil {
		return err
	}

	return nil
}

func getTwoRandomTeamsThatHaventPlayed(teams []Team) (Team, Team) {
	var team1, team2 Team
	length := len(teams)

	for i := 0; i <= length; i++ {
		rand.NewSource(time.Now().UnixNano())
		randNumOne := rand.Intn(length)
		randNumTwo := rand.Intn(length)
		team1 = teams[randNumOne]
		team2 = teams[randNumTwo]

		if team1.Name != team2.Name && !team1.Played && !team2.Played {
			break
		}
	}

	return team1, team2
}

func main() {
	server := fiber.New()
	server.Static("/", "./view/")

	server.Get("/teams", func(c *fiber.Ctx) error {
		teams, err := teamParser()
		if err != nil {
			return err
		}

		return c.JSON(teams)
	})

	server.Get("/addTeam", func(c *fiber.Ctx) error {
		team := new(Team)
		team.Name = c.Query("name")

		if err := addTeam(team); err != nil {
			c.Status(400)
			return c.SendString(err.Error())
		}

		return nil
	})

	server.Get("/startGame", func(c *fiber.Ctx) error {
		// // start the game
		game := new(Game)
		teams, err := teamParser()
		if err != nil {
			return err
		}

		team1, team2 := getTwoRandomTeamsThatHaventPlayed(teams)
		game.First_team = team1
		game.Second_team = team2
		game.Start()

		return json.NewEncoder(c).Encode(game)
	})

	server.Get("/getQuestion", func(c *fiber.Ctx) error {
		q := new(Questions)
		var err error

		q.File, err = os.OpenFile("./Teams/trivia.md", os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		q.ParseTrivia()

		question := q.GetQuestion(false)
		return json.NewEncoder(c).Encode(question)
	})

	server.Get("updateScore/:team/:round/:score", func(c *fiber.Ctx) error {
		teamName := c.Params("team")
		round, _ := strconv.Atoi(c.Params("round"))
		score, _ := strconv.Atoi(c.Params("score"))

		teams, err := teamParser()
		if err != nil {
			return err
		}

		for _, team := range teams {
			if team.Name == teamName {
				switch round {
					case 1:
						team.FirstRound += score
					case 2:
						team.SecondRound += score
					case 3:
						team.ThirdRound += score
					case 4:
						team.FourthRound += score
					
					default:
						return fmt.Errorf("Invalid Round")
				}

				team.Update()
			}
		}

		return nil
	})

	log.Fatal(server.Listen(":8080"))
}
