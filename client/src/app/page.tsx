'use client'

import { useRef, useState, useEffect } from 'react';
import styles from './page.module.css'
import mapboxgl from 'mapbox-gl';
import { Feature, Geometry, GeoJsonProperties } from 'geojson';

const geolocationStreamAPI = 'wss://map-project-backend.onrender.com/geolocation/stream';

interface GeolocationMessage {
  geolocations: Geolocation[];
}

interface Geolocation {
  device_id: string;
  event_time: Date;
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
  const [socketShouldReconnect, setSocketShouldReconnect] = useState<boolean>(true);
  const [lastPing, setLastPing] = useState<Date | null>(null);
  const [lastPong, setLastPong] = useState<Date | null>(null);
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
    if (socket.current || !socketShouldReconnect) {
      return;
    }
    console.log('Connecting to WebSocket...')

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
      if (event.data === 'pong') {
        setLastPong(new Date());
        console.log('pong');
        return;
      }

      try {
        const json: GeolocationMessage = JSON.parse(event.data, (key, value) => {
          if (key === 'event_time' && typeof value === 'string') {
            return new Date(value);
          }
          return value;
        });
        for (const geolocation of json.geolocations) {
          const lastGeolocation = geolocations.get(geolocation.device_id);
          if (!lastGeolocation) {
            // new geolocation
            geolocations.set(geolocation.device_id, geolocation);
            continue;
          }
          if (lastGeolocation.event_time > geolocation.event_time) {
            // stale geolocation
            continue;
          }
          geolocations.set(geolocation.device_id, geolocation);
        }
        setGeolocations(new Map(geolocations));
      } catch (error) {
        console.error('Error while reading WebSocket message:', error);
      }
    });

    // call socket on an interval and reconnect if needed
    const intervalId = setInterval(() => {
      const resetConnection = () => {
        console.log('WebSocket connection lost.');
        ws.close();
        socket.current = null;
        setSocketShouldReconnect(true);
        clearInterval(intervalId);
      }
      if (ws.readyState === 1) {
        if (lastPing && lastPong && lastPing > lastPong) {
          console.log('Ping timeout.');
          resetConnection();
          return;
        }
        ws.send('ping');
        setLastPing(new Date());
        console.log('ping');
      } else {
        resetConnection();
      }
    }, 5000);
    setSocketShouldReconnect(false);

    socket.current = ws;
    return () => {
      if (ws.readyState === 1) {
        clearInterval(intervalId);
        ws.close();
      }
    };
  }, [socketShouldReconnect]);

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

    // TODO: figure out what to cast this to because setData() exists
    map.current.getSource('device-locations').setData(
      {
        type: 'FeatureCollection',
        features: GeolocationsToFeatureCollection(Array.from(geolocations.values())),
      }
    );
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
