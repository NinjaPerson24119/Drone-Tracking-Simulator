CREATE OR REPLACE FUNCTION notify_on_insert_geolocation() RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify('geolocation_inserted', NEW.device_id::TEXT);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_after_insert_geolocation
AFTER INSERT ON device.geolocation
FOR EACH ROW
EXECUTE FUNCTION notify_on_insert_geolocation();
