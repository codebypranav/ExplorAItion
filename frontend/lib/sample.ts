import type { Itinerary, Place } from './types'

/**
 * Sample data for the UI's demo mode.
 *
 * The backend needs a seeded Pinecone index to answer anything; this lets the
 * interface be explored (and reviewed) without one. Anything rendered from here
 * is always labelled as sample data in the UI — see components/DemoNotice.tsx.
 */
const PARIS: Place[] = [
  {
    xid: 'sample-louvre',
    name: 'Louvre Museum',
    country: 'FR',
    description:
      'Former royal palace turned into the largest art museum in the world. Arrive early for the quieter wings.',
    score: 0.912,
    rating: 4.7,
    image_url: '',
    latitude: 48.8606,
    longitude: 2.3376,
    weather: { temperature_c: 17.4, wind_kph: 11, weather_code: 2 },
  },
  {
    xid: 'sample-orsay',
    name: "Musée d'Orsay",
    country: 'FR',
    description:
      'Impressionist collection housed in a Beaux-Arts railway station, with a famous clock window over the Seine.',
    score: 0.884,
    rating: 4.7,
    image_url: '',
    latitude: 48.86,
    longitude: 2.3266,
    weather: { temperature_c: 17.2, wind_kph: 10, weather_code: 2 },
  },
  {
    xid: 'sample-orangerie',
    name: "Musée de l'Orangerie",
    country: 'FR',
    description:
      "Small museum built around Monet's water lily panels, a short walk through the Tuileries.",
    score: 0.861,
    rating: 4.6,
    image_url: '',
    latitude: 48.8638,
    longitude: 2.3226,
    weather: { temperature_c: 17.1, wind_kph: 10, weather_code: 3 },
  },
  {
    xid: 'sample-sainte-chapelle',
    name: 'Sainte-Chapelle',
    country: 'FR',
    description:
      'Gothic chapel on the Île de la Cité whose upper level is almost entirely stained glass.',
    score: 0.847,
    rating: 4.6,
    image_url: '',
    latitude: 48.8554,
    longitude: 2.345,
    weather: { temperature_c: 16.9, wind_kph: 12, weather_code: 3 },
  },
  {
    xid: 'sample-notre-dame',
    name: 'Notre-Dame de Paris',
    country: 'FR',
    description:
      'The cathedral at the centre of the city, reopened after a long restoration of its roof and spire.',
    score: 0.833,
    rating: 4.7,
    image_url: '',
    latitude: 48.853,
    longitude: 2.3499,
    weather: { temperature_c: 16.8, wind_kph: 12, weather_code: 3 },
  },
  {
    xid: 'sample-shakespeare',
    name: 'Shakespeare and Company',
    country: 'FR',
    description:
      'English-language bookshop facing the Seine, with a quiet café serving good filter coffee next door.',
    score: 0.804,
    rating: 4.5,
    image_url: '',
    latitude: 48.8526,
    longitude: 2.3471,
    weather: { temperature_c: 16.8, wind_kph: 12, weather_code: 3 },
  },
  {
    xid: 'sample-luxembourg',
    name: 'Jardin du Luxembourg',
    country: 'FR',
    description:
      'Formal gardens with a central basin, chairs scattered everywhere and one of the best people-watching lawns in Paris.',
    score: 0.781,
    rating: 4.8,
    image_url: '',
    latitude: 48.8462,
    longitude: 2.3372,
    weather: { temperature_c: 16.5, wind_kph: 9, weather_code: 1 },
  },
  {
    xid: 'sample-rodin',
    name: 'Musée Rodin',
    country: 'FR',
    description:
      "Sculpture museum in a mansion and garden, with The Thinker set among the hedges.",
    score: 0.76,
    rating: 4.6,
    image_url: '',
    latitude: 48.8553,
    longitude: 2.3158,
    weather: { temperature_c: 16.6, wind_kph: 9, weather_code: 1 },
  },
  {
    xid: 'sample-eiffel',
    name: 'Eiffel Tower',
    country: 'FR',
    description:
      'The 1889 iron tower on the Champ de Mars. Book a timed slot, or settle for the view from Trocadéro.',
    score: 0.742,
    rating: 4.6,
    image_url: '',
    latitude: 48.8584,
    longitude: 2.2945,
    weather: { temperature_c: 16.3, wind_kph: 14, weather_code: 2 },
  },
  {
    xid: 'sample-montmartre',
    name: 'Sacré-Cœur, Montmartre',
    country: 'FR',
    description:
      'Hilltop basilica above the old artists’ quarter, with the widest free view across the rooftops.',
    score: 0.719,
    rating: 4.8,
    image_url: '',
    latitude: 48.8867,
    longitude: 2.3431,
    weather: { temperature_c: 16.1, wind_kph: 13, weather_code: 2 },
  },
  {
    xid: 'sample-canal',
    name: 'Canal Saint-Martin',
    country: 'FR',
    description:
      'Tree-lined canal with iron footbridges, coffee roasters and a long row of evening terraces.',
    score: 0.698,
    rating: 4.5,
    image_url: '',
    latitude: 48.8717,
    longitude: 2.3663,
    weather: { temperature_c: 16.4, wind_kph: 11, weather_code: 2 },
  },
  {
    xid: 'sample-pere-lachaise',
    name: 'Père-Lachaise Cemetery',
    country: 'FR',
    description:
      'Cobbled hillside cemetery that reads as a park, with cast-iron tombs under old chestnut trees.',
    score: 0.671,
    rating: 4.7,
    image_url: '',
    latitude: 48.8614,
    longitude: 2.3933,
    weather: { temperature_c: 16.5, wind_kph: 11, weather_code: 3 },
  },
]

export const SAMPLE_QUERY = '3-day trip to Paris with art museums and good coffee'

export function samplePlaces(topK = 12): Place[] {
  return PARIS.slice(0, Math.max(1, Math.min(PARIS.length, topK)))
}

/**
 * Splits the sample places into days the way the backend does: best scores
 * first, then chunked in route order.
 */
export function sampleItinerary(days = 3): Itinerary {
  const ordered = [
    'sample-louvre',
    'sample-orangerie',
    'sample-orsay',
    'sample-rodin',
    'sample-eiffel',
    'sample-luxembourg',
    'sample-shakespeare',
    'sample-notre-dame',
    'sample-sainte-chapelle',
    'sample-canal',
    'sample-montmartre',
    'sample-pere-lachaise',
  ]
    .map((xid) => PARIS.find((p) => p.xid === xid))
    .filter((p): p is Place => !!p)

  const d = Math.max(1, Math.min(14, days))
  const perDay = Math.ceil(ordered.length / d)
  const out: Itinerary = Array.from({ length: d }, () => [])
  ordered.forEach((place, i) => {
    out[Math.min(d - 1, Math.floor(i / perDay))].push(place)
  })
  return out
}
