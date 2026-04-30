CREATE TABLE IF NOT EXISTS attendance (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id     UUID NOT NULL REFERENCES staff(id),
    shop_id      UUID NOT NULL REFERENCES shops(id),
    clock_in     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    clock_out    TIMESTAMPTZ,
    shift_notes  TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_attendance_staff  ON attendance(staff_id);
CREATE INDEX IF NOT EXISTS idx_attendance_shop   ON attendance(shop_id);
CREATE INDEX IF NOT EXISTS idx_attendance_date   ON attendance(DATE(clock_in));
