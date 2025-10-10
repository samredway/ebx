# Ebiten Engine Extension (ebx)

**ebx** is my tiny, personal toolkit on top of [Ebiten](https://ebiten.org/) for making small 2D games in Go.  

I started writing this after doing some small game prototypes in Godot. While Godot is excellent, I realised I only needed a small subset of its features for my own projects—and I prefer writing code directly over working inside a heavy editor. I also wanted to stay in my favourite language (Go) and couldn’t find an existing engine that fit that niche.

So this is a super-minimal 2D engine designed for building out ideas fast.

---

## Design Philosophy

- **Scene-based architecture:**  
  A `Scene` interface with `OnEnter`, `OnExit`, `Update`, and `Draw`, plus a `Game` wrapper implementing `ebiten.Game`.  
  Scenes define their own entities, logic, and transitions, keeping state self-contained and easy to reason about.

- **Simple ECS pattern:**  
  Entities are just IDs (`EntityId` + `IdGen`).  
  Components hold data, and systems operate on those components.  
  The goal isn’t cache optimisation—it’s clean, decoupled, composable code. Each system manages its own data and can evolve independently without tight coupling to other systems.

- **System-driven updates:**  
  Instead of looping per entity, each system iterates its own component collection and performs updates or draws in a defined order.  
  This keeps logic modular, predictable, and easy to extend.

- **Small and composable:**  
  Everything is kept as small, clear packages with no global state. Each piece—scene, camera, tiles, assets—can be used on its own or combined as needed.

---

## License

Released under the **MIT License** — free to use, modify, and build on.  
Contributions, forks, and experiments are all welcome.

---

## Status

Work in progress and intentionally lightweight.  
Code is written to be idiomatic, readable, and easy to extend for quick 2D prototypes.
