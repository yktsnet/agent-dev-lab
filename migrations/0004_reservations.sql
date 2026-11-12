CREATE TABLE reservations (
  id VARCHAR(32) PRIMARY KEY,
  event_id VARCHAR(32) NOT NULL,
  attendee_id VARCHAR(32) NOT NULL,
  seats INT NOT NULL,
  status VARCHAR(16) NOT NULL,
  FOREIGN KEY (event_id) REFERENCES events(id),
  FOREIGN KEY (attendee_id) REFERENCES attendees(id)
);
