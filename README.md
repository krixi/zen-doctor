# zen-doctor
This is just a little game I'm playing around with. It's built using [gocui](https://github.com/jroimartin/gocui).
To play it, just clone this repo and run `go run cmd/main.go`

If you can't see all the symbols because your font doesn't support them, try running with `--latin`, or `--ascii` for only ASCII characters.


ideas:

1. Maps that have fixed objectives instead of completely random
2. ~footprints where you've been~
3. AI enemies: patrol (with vision cone), follow your tracks, homing on your location (unleashed on 100% threat - game over when they catch you)
4. auto-move on double-tap input
5. Items you can use - build pylons, portals, or things
6. Change the direction of the bit stream
7. Power-ups: slow down bit stream, immune to negative bits, speed up looting, multiplier on loot value, timed power up for increasing vision range, timed power up for reducing threat faster, timed "combat" power up that reduces threat when running into an enemy instead of increasing it
8. Walls and things that interrupt the bit stream - give shelter
9. Show stats at end of game for collisions with bit stream - how many and what kind of both good and bad bits
10. Pause button
11. Compatibility layer for consoles that can't display certain characters
