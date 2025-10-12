package main

import (
	"github.com/samredway/ebx/engine"
)

type ExampleScene struct {
	engine.SceneBase
}

func (es *ExampleScene) OnEnter() {
	es.SceneBase.OnEnter()

	NewPlayer(es.Ids, es.RenderSys, es.PosStore, es.MoveSys, es.UserInputSys)

	// Transform component
	// Input hanlder
	// Should now be able to see player and move around.
	// Perhpas move player into prefabs?
}
