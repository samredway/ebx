# Ebiten Engine Extension (ebx)

**ebx** is a tiny, personal toolkit on top of [Ebiten](https://ebiten.org/) for making small 2D games in Go.  

After prototyping in Godot, I realised I only needed a small subset of its features — and I prefer writing code directly without an editor. I also wanted to stay in Go and couldn’t find an existing engine that fit that niche.

So this is a minimal 2D engine for building out ideas fast.

---

## Design Philosophy

- **Scene-based architecture:**  
  A `Scene` interface with `OnEnter`, `OnExit`, `Update`, and `Draw`, plus a `Game` wrapper implementing `ebiten.Game`.  
  Scenes define their own entities, logic, and transitions, keeping state self-contained and easy to reason about.

- **Simple entity + script model:**  
  Entities are small structs holding data (e.g., Position, Movement, Render).  
  Optional scripts provide per-entity behaviour via `Update`, making gameplay logic local and easy to follow.

- **System-driven updates:**  
  Systems iterate active entities to apply shared behaviour (movement, rendering, etc).  
  Scripts run before systems, allowing gameplay code to drive animation and movement cleanly.

- **Small and composable:**  
  Everything is kept minimal, with no global state.  
  Each piece — scene, camera, tiles, assets — can be used independently.

I enjoy the composability and organisation benefits of ECS, but don't need cache-optimisation for small games. I initially built a pure ECS, but generalising every behaviour added friction — so this version uses an entity + script model for specialised logic (like AI and animations) with systems for shared mechanics.

---

## How to use

Run an example:

```bash
go mod download
go run ./examples/top-down
```
The examples show how scenes, entities, and scripts fit together.

## License

Released under the **MIT License** — free to use, modify, and build on.  
Contributions, forks, and experiments are all welcome.

---

## Status

Work in progress and intentionally lightweight.  
Code is written to be idiomatic, readable, and easy to extend for quick 2D prototypes.



