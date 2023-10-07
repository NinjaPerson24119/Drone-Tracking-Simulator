'use client'

import { useRef, useState, useEffect } from 'react';
import styles from './page.module.css'
import mapboxgl from 'mapbox-gl'; // eslint-disable-line import/no-webpack-loader-syntax

export default function Home() {
  const mapContainer = useRef<HTMLDivElement | null>(null);
  const map = useRef<mapboxgl.Map | null>(null);

  // Edmonton, Legislature
  const [longitude, setLongitude] = useState(-113.5068);
  const [latitude, setLatitude] = useState(53.5357);
  const [zoom, setZoom] = useState(16.2);

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
      setLongitude(parseFloat(map.current.getCenter().lng.toFixed(4)));
      setLatitude(parseFloat((map.current.getCenter().lat.toFixed(4))));
      setZoom(parseFloat(map.current.getZoom().toFixed(2)));
    });
  })

  return (
    <main className={styles.main}>
      <div className={styles.detailsContainer}>
        <h1>Drone Tracker</h1>
        <br/>
        <p>Longitude: {longitude}</p>
        <p>Latitude: {latitude}</p>
        <p>Zoom: {zoom}</p>
      </div>
      <div ref={mapContainer} className={styles.mapContainer}></div>
    </main>
  )
}
