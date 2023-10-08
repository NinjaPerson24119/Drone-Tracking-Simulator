'use client'

import { useRef, useState, useEffect } from 'react';
import styles from './page.module.css'
import mapboxgl from 'mapbox-gl'; // eslint-disable-line import/no-webpack-loader-syntax
import { Feature, Geometry, GeoJsonProperties } from 'geojson';

const geolocationStreamAPI = 'wss://map-project-backend.onrender.com/geolocation/stream';

interface GeolocationMessage {
  geolocations: Geolocation[];
}

interface Geolocation {
  device_id: string;
  latitude: number;
  longitude: number;
}

function GeolocationsToFeatureCollection(geolocations: Geolocation[]): Feature<Geometry, GeoJsonProperties>[] {
  return geolocations.map((geolocation) => {
    return {
      type: 'Feature',
      geometry: {
        type: 'Point',
        coordinates: [geolocation.longitude, geolocation.latitude],
      },
      properties: {
        id: geolocation.device_id,
      },
    };
  });
}

export default function Home() {
  const [geolocations, setGeolocations] = useState<Map<string, Geolocation>>(new Map<string, Geolocation>());
  const [layerAdded, setLayerAdded] = useState<boolean>(false);
  const [mapStyleLoaded, setMapStyleLoaded] = useState<boolean>(false);
  const socket = useRef<WebSocket | null>(null);

  const mapContainer = useRef<HTMLDivElement | null>(null);
  const map = useRef<mapboxgl.Map | null>(null);

  // Edmonton, Legislature
  const [latitude, setLatitude] = useState(53.5357);
  const [longitude, setLongitude] = useState(-113.5068);
  const [zoom, setZoom] = useState(16.2);

  // initialize map
  useEffect(() => {
    if (map.current) {
      return;
    }
    if (!mapContainer.current) {
      return;
    }

    const m = new mapboxgl.Map({
      container: mapContainer.current,
      center: [longitude, latitude],
      style: 'mapbox://styles/mapbox/streets-v12',
      zoom: zoom,
      accessToken: process.env.NEXT_PUBLIC_MAPBOX_API_KEY || '',
    });
    map.current = m;

    map.current.on('move', () => {
      if (!map.current) {
        return;
      }
      setLatitude(parseFloat((map.current.getCenter().lat.toFixed(4))));
      setLongitude(parseFloat(map.current.getCenter().lng.toFixed(4)));
      setZoom(parseFloat(map.current.getZoom().toFixed(2)));
    });

    map.current.on("style.load", () => {
      setMapStyleLoaded(true);
    });
  })

  // connect to websocket and listen to geolocation stream
  useEffect(() => {
    if (socket.current) {
      return;
    }

    // TODO: handle retrying connection
    const ws = new WebSocket(geolocationStreamAPI);
    ws.addEventListener('open', () => {
      console.log('WebSocket connection established.');
    });
    ws.addEventListener('close', () => {
      console.log('WebSocket connection closed.');
    });
    ws.addEventListener('error', (error) => {
      console.error('WebSocket error:', error);
    });
    ws.addEventListener('message', (event) => {
      try {
        // TODO: validate schema
        const json: GeolocationMessage = JSON.parse(event.data);
        for (const geolocation of json.geolocations) {
          geolocations.set(geolocation.device_id, geolocation);
        }
        setGeolocations(new Map(geolocations));
        //console.log('Geolocations:', geolocations);
      } catch (error) {
        console.error('Error while reading WebSocket message:', error);
      }
    });
    socket.current = ws;
    return () => {
      if (ws.readyState === 1) {
        ws.close();
      }
    };
  })

  // add/update markers as layers on map
  useEffect(() => {
    if (!map.current) {
      return;
    }
    if (!mapStyleLoaded) {
      return;
    }
    if (!layerAdded) {
      map.current.addSource('device-locations', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: GeolocationsToFeatureCollection(Array.from(geolocations.values())),
        }
      });
      map.current.addLayer({
        id: 'device-locations-layer',
        type: 'circle',
        source: 'device-locations',
        paint: {
          'circle-color': '#11b4da',
          'circle-radius': 10,
          'circle-stroke-width': 1,
          'circle-stroke-color': '#fff'
        }
      });
      setLayerAdded(true);
    }

    map.current.getSource('device-locations').setData(
      {
        type: 'FeatureCollection',
        features: GeolocationsToFeatureCollection(Array.from(geolocations.values())),
      }
    );
    //console.log('Updated map source data.');
  }, [geolocations])

  return (
    <main className={styles.main}>
      <div className={styles.detailsContainer}>
        <h1>Drone Tracker</h1>
        <br />
        <p>Latitude: {latitude}</p>
        <p>Longitude: {longitude}</p>
        <p>Zoom: {zoom}</p>
      </div>
      <div ref={mapContainer} className={styles.mapContainer}></div>
    </main>
  )
}
