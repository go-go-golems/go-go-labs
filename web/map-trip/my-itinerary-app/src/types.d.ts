interface PointOfInterest {
  name: string;
  lat: number;
  lng: number;
  bestTime: string;
  photoTips: string;
}

interface Itinerary {
  id: string;
  title: string;
  points: PointOfInterest[];
}