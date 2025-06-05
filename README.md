# lyrics

`lyrics` downlads song lyrics for all hymns in the following Church of Jesus Christ of Latter-Day Saints hymn books.
- Hymns (1985)
- Hymns-For Home and Church
- Childrens Songbook (1995)

## Usage

```
Usage of lyrics:
  -collection string
    	path to the collection (will be created if missing) (default "collection")
```

## Collection

All of the hymns are written to (by default) `./collection/<book>/songs/`.  Each lyric file is then symlinked to `./collection/<book>/<section>/` as a simple means of organization.
