package main

type Game struct {
	First_team  Team
	Second_team Team
	Questions   Questions
}

func (g *Game) Setup() {}

func (g *Game) Start() {}

func (g *Game) End() {}
