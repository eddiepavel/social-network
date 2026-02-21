package contextkeys

import "time"

const UserIDKey = "user_id"
const SessionCookieName = "session_id"
const SessionDuration = 7 * 24 * time.Hour // 7 days
const GuestSession = "guest_session"
