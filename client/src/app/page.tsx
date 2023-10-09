'use client'

import { useRef, useState, useEffect } from 'react';
import styles from './page.module.css'
import mapboxgl from 'mapbox-gl';
import { Feature, Geometry, GeoJsonProperties } from 'geojson';

const geolocationStreamAPI = process.env.NEXT_PUBLIC_WEBSOCKET || '';

interface GeolocationMessage {
  geolocations: Geolocation[];
}

interface Geolocation {
  device_id: string;
  event_time: Date;
  latitude: number;
  longitude: number;
}

interface GeolocationCluster {
  avgLatitude: number;
  avgLongitude: number;
  geolocations: Geolocation[];
}

function GeolocationsToClusters(geolocations: Map<string, Geolocation>, clusterDistance: number): Array<GeolocationCluster> {
  const maxDistSquared = Math.pow(clusterDistance, 2);
  const clusters = new Array<GeolocationCluster>();
  for (const geolocation of geolocations.values()) {
    let nearestDistance = 0;
    let nearestCluster: GeolocationCluster | null = null;
    // iterate clusters to find the nearest one
    for (const cluster of clusters) {
      const distSquared = 
        Math.pow(cluster.avgLatitude - geolocation.latitude, 2) +
        Math.pow(cluster.avgLongitude - geolocation.longitude, 2)
      if (distSquared > maxDistSquared) {
        continue;
      }
      nearestDistance = distSquared;
      nearestCluster = cluster;
    }
    // if nothing is within the cluster distance, create a singleton cluster
    if (nearestCluster == null) {
      
      clusters.push({
        avgLatitude: geolocation.latitude,
        avgLongitude: geolocation.longitude,
        geolocations: [geolocation],
      });
    } else {
      // otherwise, add the geolocation to the nearest cluster
      nearestCluster.geolocations.push(geolocation);
      nearestCluster.avgLatitude = (nearestCluster.avgLatitude + geolocation.latitude) / 2;
      nearestCluster.avgLongitude = (nearestCluster.avgLongitude + geolocation.longitude) / 2;
    }
  }
  return clusters;
}

function GeolocationClustersToFeatureCollection(clusters: GeolocationCluster[]): Feature<Geometry, GeoJsonProperties>[] {
  return clusters.map((cluster, idx) => {
    return {
      type: 'Feature',
      geometry: {
        type: 'Point',
        coordinates: [cluster.avgLongitude, cluster.avgLatitude],
      },
      properties: {
        id: `cluster-${idx}`,
      },
    };
  });
}

export default function Home() {
  const [geolocations, setGeolocations] = useState<Map<string, Geolocation>>(new Map<string, Geolocation>());
  const [geolocationClusters, setGeolocationClusters] = useState<Map<string, GeolocationCluster>>(new Map<string, GeolocationCluster>());
  const [layerAdded, setLayerAdded] = useState<boolean>(false);
  const [mapStyleLoaded, setMapStyleLoaded] = useState<boolean>(false);
  const [socketShouldReconnect, setSocketShouldReconnect] = useState<boolean>(true);
  const [lastPing, setLastPing] = useState<Date | null>(null);
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
    const sendPing = () => {
      setLastPing(new Date());
      ws.send('ping');
    }
    ws.addEventListener('open', () => {
      console.log('WebSocket connection established.');
    });
    ws.addEventListener('close', () => {
      console.log('WebSocket connection closed.');
      resetConnection();
    });
    ws.addEventListener('error', (error) => {
      console.error('WebSocket error:', error);
      resetConnection();
    });
    ws.addEventListener('message', (event) => {
      if (event.data === 'pong') {
        setTimeout(sendPing, 5000);
        return;
      }

      try {
        const json: GeolocationMessage = JSON.parse(event.data, (key, value) => {
          if (key === 'event_time' && typeof value === 'string') {
            return new Date(value);
          }
          return value;
        });
        console.log('Received geolocations');
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
          // update geolocation
          geolocations.set(geolocation.device_id, geolocation);
        }
        setGeolocations(new Map(geolocations));
      } catch (error) {
        console.error('Error while reading WebSocket message:', error);
      }
    });

    // call socket on an interval and reconnect if needed
    // browser will not send pings so we need to do this manually
    const intervalId = setInterval(() => {
      ws.dispatchEvent(new Event('checkPing'));
    }, 6000);
    ws.addEventListener('checkPing', () => {
      if (ws.readyState !== 1) {
        console.log('WebSocket not ready.');
        return;
      }
      if (!lastPing) {
        sendPing();
        return;
      }
      const pingElapsed = new Date().getTime() - lastPing.getTime();
      if (pingElapsed > 7000) {
        console.log('Ping timeout.');
        resetConnection();
        return;
      }
    });

    setSocketShouldReconnect(false);
    const resetConnection = () => {
      console.log('WebSocket connection lost.');
      ws.close();
      socket.current = null;
      setSocketShouldReconnect(true);
      setLastPing(null);
      clearInterval(intervalId);
    }
    socket.current = ws;
    return () => {
      if (ws.readyState === 1) {
        ws.close();
        clearInterval(intervalId);
      }
    };
  });

  // add/update markers as layers on map
  useEffect(() => {
    if (!map.current) {
      return;
    }
    if (!mapStyleLoaded) {
      return;
    }

    // TODO: linearly relate zoom to the cluster distance
    const clusterDistance = 0.01;
    const clusters = GeolocationsToClusters(geolocations, clusterDistance);

    if (!layerAdded) {
      map.current.addSource('device-locations', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: GeolocationClustersToFeatureCollection(clusters),
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
        },
        layout: {
          'text-field': ['get', 'point_count_abbreviated'],
          'text-font': ['DIN Offc Pro Medium', 'Arial Unicode MS Bold'],
          'text-size': 12
          }
      });
      setLayerAdded(true);
    }

    

    // TODO: figure out what to cast this to because setData() exists
    map.current.getSource('device-locations').setData(
      {
        type: 'FeatureCollection',
        features: GeolocationClustersToFeatureCollection(clusters),
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
