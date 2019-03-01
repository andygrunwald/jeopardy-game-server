# jeopardy-game-server

## Usage

### Start

```sh
$ make run
```

### Env vars

This environment vars are available:

| Name             | Default       | Description                               |
| ---------------- | ------------- | ----------------------------------------- |
| JGAME_SRV_ADDR   | :8000         | IP + Port where the webserver will listen |

## Shipping a new release

1. Make your modifications => `hack hack hack`
2. Create a new tag => `git tag -a v0.1.0 -m "New release" && git push origin v0.1.0`
3. Release it => `goreleaser`

For more, check out the [goreleaser](https://goreleaser.com/) docs.

## Games

All games are stored in the `games/` folder.
Inside this folder, there can be one up to n _Seasons_.
One _Season_ can contain one up to m _Games_.
One _Game_ contains one Jeopardy! Game.

The folder/file structure can look like this:

```
├── games
│   └── example-season-1
│       ├── 2019-02-14
│       │   ├── game.json
│       │   └── overview.json
│       ├── 2019-02-25
│       │   ├── game.json
│       │   └── overview.json
│       ├── [...]
│       └── overview.json
```

Have a look at [games/Season_1](./games/Season_1) for an example season + games.

### Adding a new _Season_

1. Add a new folder with a name of your choice. The name should not contain whitespaces.
   
   ```sh
   $ mkdir games/my-new-funny-event
   ```

   The folder name is not the season name. This name can be set separately.
   The folder name will be used for sorting.

2. Add a file named `overview.json` in the season folder and add the following content:

   ```json
   {
      "id": "my-new-funny-event",
      "name": "Funny Event (March 2019 in DUS)",
      "description": "Put a useful description here. Maybe Location, date or topic of the game.",
      "note": "A random note you can put here"
   }
   ```

### Adding a new _Game_ in an existing _Season_

1. Add a new folder with a name of your choice. The name should not contain whitespaces.
   
   ```sh
   $ mkdir games/my-new-funny-event/game-1
   ```

   The folder name is not the game name. This name can be set separately.
   The folder name will be used for sorting.

2. Add a file named `overview.json` in the game folder and add the following content:

   ```json
   {
      "id": "my-new-funny-event---game-1",
      "name": "#7929, aired 2019-02-14",
      "description": "Eric R. Backes vs. Alex Miller Murphy vs. Mitch Rodricks",
      "note": ""
   }
   ```

   - The `id` field has the structure of `SEASON-FOLDER---GAME-FOLDER`. The `---` is a replacement for the `/` char
  
3. Add a file named `game.json` in the game folder and add the following content:

   ```json
   {
      "id": "my-new-funny-event/game-1",
      "game_title": "Show #7929 - Thursday, February 14, 2019",
      "game_comments": "",
      "game_complete": true,
      "category_J_1": {
        "category_name": "THE TOP 25 ROMANTIC COMEDIES",
        "category_comments": "(Alex: According to a list put together by Vanity Fair.)",
        "clue_count": 5
      },
      "category_J_2": {
        [...]
      },
      [...]
      "clue_J_1_1": {
        "id": "359675",
        "clue_html": "He stars in No. 4, No. 7 &amp; No. 16: <br>&quot;Bridget Jones&apos;s Diary&quot;, &quot;Notting Hill&quot; &amp; &quot;Four Weddings and a Funeral&quot;",
        "clue_text": "He stars in No. 4, No. 7 & No. 16: \"Bridget Jones's Diary\", \"Notting Hill\" & \"Four Weddings and a Funeral\"",
        "correct_response": "Hugh Grant"
      },
      "clue_J_2_1": {
        "id": "359682",
        "clue_html": "No bull--we&apos;re stuck on this brand whose mascot is seen <a href=\"http://www.j-archive.com/media/2019-02-14_J_11.jpg\" target=\"_blank\">here</a>",
        "clue_text": "No bull--we're stuck on this brand whose mascot is seen here",
        "correct_response": "Elmer\\'s Glue",
        "media": [
            "http://localhost:3000/media/2019-02-14_J_11.jpg"
        ]
      },
      "clue_J_2_4": {
        "id": "359680",
        "daily_double": true,
        "clue_html": "Napoleon was heard to say that this, which he was part of, is merely &quot;a fable agreed upon&quot;",
        "clue_text": "Napoleon was heard to say that this, which he was part of, is merely \"a fable agreed upon\"",
        "correct_response": "history"
      },
      [...]
   ```

   - `category_J_<CAT_ID|1..6>`: A single category. Six categories can be placed there.
   - `clue_J_<CAT_ID>_<QUESTON_ID|1..5>`: A single question. `clue_J_1_2` will be in category 1 question 2

This is how a normal game look like.
Additionally you can add a Double Jeopardy! and Final Jeopardy! Round.

#### Double Jeopardy! Round

Double Jeopardy! Rounds apply the same scheme as normal Categories and Questions.
The difference is that in the JSON Key the `J` will be replaced with a `DJ`.
Like

```json
    [...]
    "category_DJ_6": {
        "category_name": "TV REBOOTED",
        "category_comments": "",
        "clue_count": 5
    },
    "clue_DJ_1_1": {
        "id": "359705",
        "clue_html": "Now in their 40s, this title trio returns to fight for justice in Dumas&apos; &quot;Twenty Years After&quot;",
        "clue_text": "Now in their 40s, this title trio returns to fight for justice in Dumas' \"Twenty Years After\"",
        "correct_response": "the Three Musketeers"
    },
    [...]
```

#### Final Jeopardy! Round

Only one final Jeopardy! round can be added.
Just add the following into the `game.json` file:

```json
    "category_FJ_1": {
        "category_name": "COLORFUL GEOGRAPHY",
        "category_comments": ""
    },
    "clue_FJ": {
        "clue_html": "Named for a soldier killed in 1846 at the start of a war, it was in the news again as a port of entry to the U.S. in 2018",
        "clue_text": "Named for a soldier killed in 1846 at the start of a war, it was in the news again as a port of entry to the U.S. in 2018",
        "correct_response": "Brownsville"
    }
```
