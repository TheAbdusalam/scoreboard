package main

type Game struct {
	First_team  Team
	Second_team Team
	Questions   Questions
}

func (g *Game) Start() {
	// g.First_team.Played = true
	// g.Second_team.Played = true

	g.First_team.Update()
	g.Second_team.Update()
}

func (g *Game) End() {}
