'use client'

import { useRef, useState, useEffect } from 'react';
import styles from './page.module.css'
import { GetServerSideProps } from 'next';
import mapboxgl from 'mapbox-gl'; // eslint-disable-line import/no-webpack-loader-syntax

interface HomeProps {
  latitude: number;
  longitude: number;
  zoom: number;
  mapboxAPIKey: string;
}

export default function Home(props: HomeProps) {
  const mapContainer = useRef<HTMLDivElement | null>(null);
  const map = useRef<mapboxgl.Map | null>(null);

  // Edmonton, Legislature
  const [longitude, setLongitude] = useState(props.latitude);
  const [latitude, setLatitude] = useState(props.longitude);
  const [zoom, setZoom] = useState(props.zoom);

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
      accessToken: props.mapboxAPIKey,
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

export const getServerSideProps: GetServerSideProps<HomeProps> = async () => {
  // Edmonton, Legislature
  const initialLongitude = -113.5068;
  const initialLatitude = 53.5357;
  const initialZoom = 16.2;

  return {
    props: {
      latitude: initialLatitude,
      longitude: initialLongitude,
      zoom: initialZoom,
      mapboxAPIKey: process.env.MAPBOX_API_KEY || '',
    },
  };
}
