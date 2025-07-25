# linjante

An API that randomly generates Toki Pona sentences. (My first API written in Go)

All data for words is from [sona Linku](https://github.com/lipu-linku/sona).

## Installation

1. [Install Go](https://go.dev/doc/install)

2. Clone this repository

   ```bash
   git clone https://github.com/malkicodes/linjante.git
   ```

3. Clone [sona Linku](https://github.com/lipu-linku/sona) into `./sona`

   ```bash
   git clone https://github.com/lipu-linku/sona.git
   ```

4. Run the server

   ```bash
   go mod tidy
   go run .
   ```

## Usage

Use the `/gen` endpoint to generate sentences!

```json
/gen

{
    "count": 1,
    "sentences": [
        "moku su li ijo e ona"
    ]
}
```

Use the `count` parameter to specify how many sentences you want to generate (up to 50 at a time)

```json
/gen?count=5

{
    "count": 5,
    "sentences": [
        "alasa tomo li ike",
        "pini pan li sona ala lawa e mi",
        "mu li wan ala lon jelo",
        "linja jelo li telo e ona lon pali",
        "nasa ku li tu lon ni tan ken"
    ]
}
```

Make the responses more verbose using the `v` parameter and give some information on the sentence structure.

```json
/gen?v=true

{
    "count": 1,
    "sentences": [
        {
            "sentence": "majuna monsuta li moli e ni lon pali suno tawa ni"
            "components": [
                "majuna monsuta",
                "moli",
                "ni",
                "lon pali suno",
                "tawa ni"
            ],
            "roles": {
                "object": "ni",
                "prepositions": [
                    "lon pali suno",
                    "tawa ni"
                ],
                "subject": "majuna monsuta",
                "verb": "moli"
            },
        }
    ]
}
```
