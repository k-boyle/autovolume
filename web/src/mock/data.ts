import { MediaType } from '../gen/autovolume/v1/media_pb'

export interface CatalogEntry {
  ratingKey: string
  type: MediaType
  title: string
  year: number
  libraryId: string
  show?: { ratingKey: string; title: string; seasonNumber: number }
}

const movies: [string, number][] = [
  ['Aftersun', 2022], ['Alien', 1979], ['Amélie', 2001], ['Arrival', 2016], ['Back to the Future', 1985],
  ['Before Sunrise', 1995], ['Blade Runner 2049', 2017], ['Booksmart', 2019], ['Casablanca', 1942],
  ['Children of Men', 2006], ['Chinatown', 1974], ['Coco', 2017], ['Drive', 2011], ['Dune', 2021],
  ['Dune: Part Two', 2024], ['Everything Everywhere All at Once', 2022], ['Ex Machina', 2014], ['Fargo', 1996],
  ['Get Out', 2017], ['Gravity', 2013], ['Heat', 1995], ['Her', 2013], ['Inception', 2010], ['Interstellar', 2014],
  ['Jaws', 1975], ['Jurassic Park', 1993], ['Knives Out', 2019], ['La La Land', 2016], ['Lost in Translation', 2003],
  ['Mad Max: Fury Road', 2015], ['Moonlight', 2016], ['No Country for Old Men', 2007], ['Oppenheimer', 2023],
  ['Paddington 2', 2017], ['Parasite', 2019], ['Past Lives', 2023], ['Prisoners', 2013], ['Ratatouille', 2007],
  ['Rear Window', 1954], ['Roma', 2018], ['Sicario', 2015], ['Spider-Man: Into the Spider-Verse', 2018],
  ['Spirited Away', 2001], ['The Dark Knight', 2008], ['The Godfather', 1972], ['The Grand Budapest Hotel', 2014],
  ['The Handmaiden', 2016], ['The Matrix', 1999], ['The Prestige', 2006], ['The Social Network', 2010],
  ['The Thing', 1982], ['There Will Be Blood', 2007], ['Top Gun: Maverick', 2022], ['Up', 2009],
  ['WALL·E', 2008], ['Whiplash', 2014], ['Zodiac', 2007],
]

const shows: [string, number, number][] = [
  ['Andor', 2022, 2], ['Arcane', 2021, 2], ['Atlanta', 2016, 4], ['Barry', 2018, 4], ['Better Call Saul', 2015, 6],
  ['Bluey', 2018, 3], ['Breaking Bad', 2008, 5], ['Chernobyl', 2019, 1], ['Dark', 2017, 3], ['Fleabag', 2016, 2],
  ['Hacks', 2021, 4], ['Mad Men', 2007, 7], ['Mr. Robot', 2015, 4], ['Only Murders in the Building', 2021, 5],
  ['Planet Earth II', 2016, 1], ['Severance', 2022, 2], ['Shōgun', 2024, 1], ['Slow Horses', 2022, 5],
  ['Succession', 2018, 4], ['Taskmaster', 2015, 19], ['Ted Lasso', 2020, 3], ['The Bear', 2022, 4],
  ['The Expanse', 2015, 6], ['The Last of Us', 2023, 2], ['The Leftovers', 2014, 3], ['The Office', 2005, 9],
  ['The Wire', 2002, 5], ['True Detective', 2014, 4], ['Twin Peaks', 1990, 3], ['What We Do in the Shadows', 2019, 6],
]

const showsWithSpecials = new Set(['Bluey', 'Taskmaster'])

export const libraries = [
  { id: '1', title: 'Movies', type: MediaType.MOVIE },
  { id: '2', title: 'TV Shows', type: MediaType.SHOW },
]

export const catalog: CatalogEntry[] = [
  ...movies.map(([title, year], i) => ({ ratingKey: `m${i + 1}`, type: MediaType.MOVIE, title, year, libraryId: '1' })),
  ...shows.map(([title, year], i) => ({ ratingKey: `s${i + 1}`, type: MediaType.SHOW, title, year, libraryId: '2' })),
]

export const seasons: CatalogEntry[] = shows.flatMap(([title, , count], i) => {
  const show = { ratingKey: `s${i + 1}`, title }
  const first = showsWithSpecials.has(title) ? 0 : 1
  return Array.from({ length: count + 1 - first }, (_, j) => {
    const seasonNumber = first + j
    return {
      ratingKey: `${show.ratingKey}-${seasonNumber}`,
      type: MediaType.SEASON,
      title: seasonNumber === 0 ? 'Specials' : `Season ${seasonNumber}`,
      year: 0,
      libraryId: '2',
      show: { ...show, seasonNumber },
    }
  })
})

export function findByTitle(title: string): CatalogEntry {
  const entry = catalog.find((e) => e.title === title)
  if (!entry) throw new Error(`mock catalog has no "${title}"`)
  return entry
}

export function findSeason(showTitle: string, seasonNumber: number): CatalogEntry {
  const entry = seasons.find((e) => e.show?.title === showTitle && e.show.seasonNumber === seasonNumber)
  if (!entry) throw new Error(`mock catalog has no "${showTitle}" season ${seasonNumber}`)
  return entry
}

export const servers = [
  { id: 'srv-home', name: 'Basement NAS', owned: true, online: true },
  { id: 'srv-shared', name: "Sam's Plex", owned: false, online: true },
]

export const players = [
  { id: 'player-tv', name: 'BRAVIA 4K UR3', product: 'Plex for Android (TV)', lastSeenMinutesAgo: 3 },
  { id: 'player-phone', name: 'Pixel 9', product: 'Plex for Android (Mobile)', lastSeenMinutesAgo: undefined },
]

export const initialRules: Record<string, number> = {
  [findByTitle('Severance').ratingKey]: 18,
  [findSeason('Severance', 2).ratingKey]: 24,
  [findByTitle('Bluey').ratingKey]: 12,
  [findByTitle('Dune: Part Two').ratingKey]: 30,
  [findByTitle('Planet Earth II').ratingKey]: 25,
}
