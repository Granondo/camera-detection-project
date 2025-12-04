-- Migration: Add frame_id to events table
-- This connects events to the specific frame that triggered them

-- Add frame_id column to events table
ALTER TABLE events
ADD COLUMN frame_id INTEGER REFERENCES frames(id) ON DELETE CASCADE;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_events_frame_id ON events(frame_id);

-- Add comment to document the relationship
COMMENT ON COLUMN events.frame_id IS 'Reference to the frame that triggered this event (nullable for system events)';
