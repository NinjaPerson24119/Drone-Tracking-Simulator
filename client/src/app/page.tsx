'use client'

import { useRef, useState, useEffect } from 'react';
import styles from './page.module.css'
import mapboxgl from 'mapbox-gl'; // eslint-disable-line import/no-webpack-loader-syntax

const geolocationStreamAPI = 'wss://map-project-backend.onrender.com/geolocation/stream';

interface GeolocationMessage {
  geolocations: Geolocation[];
}

interface Geolocation {
  device_id: string;
  latitude: number;
  longitude: number;
}

export default function Home() {
  const [geolocations, setGeolocations] = useState<Map<string, Geolocation>>(new Map<string, Geolocation>());
  const [markers, setMarkers] = useState<Map<string, mapboxgl.Marker>>(new Map<string, mapboxgl.Marker>());
  const socket = useRef<WebSocket | null>(null);

  const mapContainer = useRef<HTMLDivElement | null>(null);
  const map = useRef<mapboxgl.Map | null>(null);

  // Edmonton, Legislature
  const [latitude, setLatitude] = useState(53.5357);
  const [longitude, setLongitude] = useState(-113.5068);
  const [zoom, setZoom] = useState(16.2);

  // initialize map
  useEffect(() => {
    if(map.current) {
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
        console.log('Geolocations:', geolocations);
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

  // add/update markers
  useEffect(() => {
    if (!map.current) {
      return;
    }

    for (const geolocation of geolocations.values()) {
      if (markers.has(geolocation.device_id)) {
        const marker = markers.get(geolocation.device_id);
        if (marker) {
          marker.setLngLat([geolocation.longitude, geolocation.latitude]);
        }
        continue;
      }
      const marker = new mapboxgl.Marker({
        color: '#FF0000',
        draggable: false,
        anchor: 'center',
        //element: ReactDOM.render(() => <div className={styles.marker}></div>),
      }).setLngLat([geolocation.longitude, geolocation.latitude]).addTo(map.current);
      marker.setLngLat
    }

    setMarkers(new Map(markers));
  }, [geolocations])

  return (
    <main className={styles.main}>
      <div className={styles.detailsContainer}>
        <h1>Drone Tracker</h1>
        <br/>
        <p>Latitude: {latitude}</p>
        <p>Longitude: {longitude}</p>
        <p>Zoom: {zoom}</p>
        <p>{JSON.stringify(geolocations)}</p>
      </div>
      <div ref={mapContainer} className={styles.mapContainer}></div>
    </main>
  )
}
