# Game design

## Fantasy

Battle Bunny Wealth is a military-comedy world populated by rabbits who are chronically underfunded and wildly overconfident. The player commands a small unit that builds a warren economy to fund equipment, bizarre businesses, and bomb-arena operations.

The tone is cartoon military parody, not realistic warfare. Explosions are arcade tools. Characters are toy-like, expressive, and absurd.

## Pillars

1. **Idle progress feels good.** Open the app, collect meaningful progress, make a few decisions, leave.
2. **Active play is skill-based.** Arena victories come from movement, timing, prediction, map control, and matchup knowledge.
3. **Characters carry the story.** First Sergeant Hard-as-Nails, Da Champ, Private Stuffy, Corporal Boomboom, Captain Cashmere, Doc Flopsy, and future recruits drive chapters.
4. **The blockchain is invisible until useful.** Nobody should need to understand epochs to enjoy the game.

## Core loop

```text
idle operations -> cash/scrap -> buildings/research -> bombs/perks
       ^                                              |
       |                                              v
story/progression <- loot/rank <- grid bomb battles <-+
```

## Idle layer

Candidate operations:

- Questionable Carrot Logistics
- Scrap Recovery Detail
- Tactical Lawn Services
- Surplus Snack Depot
- Warren Energy Drink Company
- Burrow Freight & Salvage
- Officer's Club Gift Shop

Each operation has a manager/squad member, level, income rate, upgrade tree, and story beats.

## Battle layer

The battle mode is an original grid-bomb arena inspired by the readability and tension of classic maze bomb games, not a clone of another game's characters/assets/maps.

Target rules:

- 2–8 players depending on mode.
- Destructible and indestructible terrain.
- Bomb placement with readable fuse timing.
- Orthogonal blast propagation blocked by hard terrain.
- Pickups alter bomb count, radius, movement, and abilities.
- Character abilities create tactical variation without invalidating fundamentals.
- Short matches suitable for mobile.

Candidate bombs: Cherry Bomb, Carrot Charge, Trashcan TNT, Bee Bomb, Magnet Bomb, Burrow Bomb, Freeze Bomb, Glitter Bomb.

## Initial squad

- **First Sergeant Hard-as-Nails** — discipline/production; intimidating mentor.
- **Da Champ** — arena specialist; flashy and legitimately talented.
- **Private Stuffy** — rookie/scavenger; anxious but lucky.
- **Corporal Boomboom** — demolitions; enthusiastic safety hazard.
- **Captain Cashmere** — logistics/finance; profit-minded strategist.
- **Doc Flopsy** — recovery/support; exhausted professional.

## Progression

Use chapters/locations rather than endless anonymous levels:

1. Broken Burrow
2. Scrap Row
3. Carrot District
4. Downtown Warren
5. Officer Country
6. Billionaire Burrows
7. future absurd theaters

The first playable MVP should prove one idle operation, one upgrade choice, one squad member progression path, and one local arena map.
