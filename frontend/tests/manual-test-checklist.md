# Sync Chat App — Manual UI Test Checklist

> **Purpose:** Comprehensive manual test plan for Chrome-based verification of chat features, including message alignment, real-time updates, unread badges, and sidebar behavior.
> **Last verified:** May 28, 2026
> **Tested by:** AI-assisted browser-use + manual verification

---

## Prerequisites

- [ ] Backend running on `http://localhost:8080` (health check: `curl http://localhost:8080/health` → `{"status":"ok"}`)
- [ ] Frontend running on `http://localhost:3000` (navigate in Chrome)
- [ ] Chrome DevTools **Console** open — monitor for errors throughout testing
- [ ] Two browser windows / incognito sessions recommended for multi-user tests

---

## 1. Authentication Flow

### 1.1 Registration
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 1.1 | Register new user | 1. Navigate to `/register`<br>2. Fill username, email, password, confirm password<br>3. Click "Create Account" | Redirected to `/chat`; sidebar shows username |
| 1.2 | Form validation — empty | 1. Go to `/register`<br>2. Click "Create Account" with empty fields | Browser validation prevents submission; stays on `/register` |
| 1.3 | Form validation — short password | 1. Go to `/register`<br>2. Enter valid username/email, password "123"<br>3. Submit | Error shown: "at least 6 characters" |
| 1.4 | Register duplicate email | 1. Try registering with an existing email | Error message shown; stays on register page |

### 1.2 Login
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 1.5 | Login with valid credentials | 1. Navigate to `/login`<br>2. Enter email + password<br>3. Click "Sign In" | Redirected to `/chat`; "Connected" status shown |
| 1.6 | Login with invalid credentials | 1. Enter wrong email/password<br>2. Submit | Error shown; stays on `/login` |

### 1.3 Navigation
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 1.7 | Login → Register link | Click "Create one" on login page | Navigates to `/register` |
| 1.8 | Register → Login link | Click "Sign in" on register page | Navigates to `/login` |

---

## 2. Sidebar Behavior

### 2.1 User Info & Status
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 2.1 | User info in footer | After login, look at sidebar bottom | Username displayed; green dot = "Connected" |
| 2.2 | Profile dialog | Click on user avatar in sidebar footer | Profile dialog opens with user details |
| 2.3 | Logout | Click logout button (door icon) in sidebar footer | Redirected to `/login` |

### 2.2 Conversation List
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 2.4 | Empty state | New user with no conversations | "No conversations yet" message |
| 2.5 | Conversation tabs | Click "All", "1-1 Chats", "Groups" tabs | List filters by conversation type |
| 2.6 | Conversation selection | Click a conversation | Chat area opens; active conv highlighted |

### 2.3 Real-time Sidebar Updates
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 2.7 | New message moves conv to top | User A sends message to User B | In User B's sidebar, conversation jumps to top of list |
| 2.8 | Last message preview updates | User A sends "Hello there!" | User B's sidebar shows "Hello there!" as last message |
| 2.9 | Timestamp updates | User A sends a message | User B's sidebar shows updated timestamp |

---

## 3. Message Alignment

### 3.1 Own Messages (Right Side)
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 3.1 | Own text message | 1. Open a conversation<br>2. Type a message<br>3. Send | Message bubble appears on the **RIGHT** side (`justify-end`) |
| 3.2 | Own message style | Verify own message bubble | Gradient background (`from-[var(--primary)] to-[var(--accent)]`), white text, rounded corners with `rounded-tr-md` |
| 3.3 | Own message checkmark | Verify own message has ✓ or ✓✓ indicator | Single ✓ (delivered) or double ✓✓ (seen), bottom-right of bubble |

### 3.2 Other Users' Messages (Left Side)
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 3.4 | Other user's message | Log in as User B; User A's message appears | Message bubble on the **LEFT** side (`justify-start`) |
| 3.5 | Other user's message style | Verify other's bubble | Solid background (`bg-[var(--surface-2)]`), border, `rounded-tl-md` |

### 3.3 Group Messages
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 3.6 | Group: own message with "You" label | Send message in group | Shows "You" label above bubble; right-aligned |
| 3.7 | Group: other user with username | Other user sends in group | Shows sender's username above bubble; left-aligned |

---

## 4. Unread Message Badges

### 4.1 Unread Count Display
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 4.1 | Unread badge appears | User A sends message to User B; User B is in different tab/not viewing conversation | User B's sidebar shows a **gradient badge** (pink-purple gradient circle) with the unread count |
| 4.2 | Badge text formatting | Unread count 1-99 shown as number | Badge shows the exact number |
| 4.3 | Badge overflow (99+) | Simulate 100+ unread messages | Badge shows "99+" |

### 4.2 Unread States
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 4.4 | Bold conversation name | Conversation has unread messages | Conversation name is **bold** (`font-semibold`) and brighter text |
| 4.5 | Last message bold | Conversation has unread messages | Last message preview text is prominent (`font-medium`) |
| 4.6 | No unread — normal style | After clearing unread | Normal text weight; no badge |

### 4.3 Clear on Select
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 4.7 | Unread clears on click | Click conversation with unread badge | Badge disappears; count resets to 0 |
| 4.8 | Sender's own messages don't increment | User A sends message while viewing conversation A↔B | User A's conversation with B does NOT get unread badge for own message |

### 4.4 Real-time Increment
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 4.9 | Unread increments live | While User B has conversation A↔B inactive, User A sends multiple messages | User B's unread badge increments with each message |

---

## 5. Chat Page Layout & Scrolling

### 5.1 Chat Container
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 5.1 | Empty state | No conversation selected | "Select a Conversation" placeholder shown |
| 5.2 | Messages visible | Open conversation with messages | Messages fill the middle area of the page |
| 5.3 | Message input visible | At bottom of chat | Input field with placeholder "Type a message..." always visible |

### 5.2 Scrolling
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 5.4 | Scroll when many messages | Send enough messages to overflow the viewport | Chat area becomes **scrollable** with a scrollbar |
| 5.5 | Auto-scroll to bottom | New message received | Automatically scrolls to bottom to show latest message |
| 5.6 | Input stays visible during scroll | Scroll up, then back down | Input bar never gets hidden or pushed off-screen |

### 5.3 Date Separators
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 5.7 | Date separator appears | Messages from different days | "Today", "Yesterday", or date string separator between groups |
| 5.8 | No duplicate separators | Multiple messages same day | Only one separator for the day |

---

## 6. Real-time Messaging

### 6.1 Message Delivery
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 6.1 | Optimistic send | User A sends message | Message appears instantly (optimistic UI) before API returns |
| 6.2 | Real-time receive | User B receives message without refresh | Message appears in User B's chat in real-time via WebSocket |
| 6.3 | No duplicate keys | Send multiple messages rapidly | No React console errors about duplicate keys |
| 6.4 | File send | Click paperclip, select a file, send | File attachment appears with preview for images |

### 6.2 Message Read Receipts
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 6.5 | Read receipt sent | User B opens conversation with new messages | Read receipt WebSocket message sent |
| 6.6 | ✓✓ seen indicator | User A's message shows double check | After User B sees message, User A sees ✓✓ (blue) instead of ✓ (gray) |

### 6.3 Reactions
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 6.7 | Add reaction | Click emoji button on a message, select an emoji | Emoji appears below the message |
| 6.8 | Toggle reaction | Click same emoji again | Emoji removed |
| 6.9 | Count display | Multiple users react with same emoji | Count shown next to emoji |
| 6.10 | Real-time reaction | User B reacts; User A sees it without refresh | Reaction appears live on User A's screen |

---

## 7. Search & Discovery

### 7.1 Message Search
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 7.1 | Open search | Click search icon in chat header | Search input opens; messages area dims |
| 7.2 | Search results | Type a word that exists in chat messages | Matching messages shown with highlighted text |
| 7.3 | Navigate to result | Click a search result | Scrolls to that message; highlights it briefly |
| 7.4 | Close search | Press Escape or click back arrow | Search closes; normal view returns |

### 7.2 Discovery Dialog
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 7.5 | Open discovery | Click compass/search icon in sidebar header | Discovery dialog opens |
| 7.6 | User search | Type a username in discovery search | Matching users appear |
| 7.7 | Start conversation | Click on a search result user | Conversation opens in main chat area |

### 7.3 Group Creation
| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 7.8 | Open create group | Click "+" button in sidebar footer | Create Group dialog opens |
| 7.9 | Add members | Search for users; click to add | Users appear as selected chips |
| 7.10 | Create group | Enter name, click "Create Group" | Group appears in sidebar; chat opens |
| 7.11 | Send in group | Type message and send | All group members see the message in real-time |

---

## 8. Loading & Error States

| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 8.1 | Loading spinner (conversations) | Slow network; observe initial load | Spinner shown while conversations fetch |
| 8.2 | Loading spinner (messages) | Click conversation with many messages | Spinner during message loading |
| 8.3 | Backend down | Stop backend; try to use app | Error message or offline state handled gracefully |
| 8.4 | WebSocket reconnection | Stop backend; restart it | Reconnects automatically; shows "Connected" |

---

## 9. Accessibility (Console Errors)

| # | Test | Steps | Expected Result | Pass/Fail |
|---|------|-------|-----------------|-----------|
| 9.1 | No React key warnings | Send several messages quickly | Console: no "two children with the same key" errors |
| 9.2 | No useEffect size warnings | Navigate between conversations | Console: no "useEffect changed size between renders" warnings |
| 9.3 | Form labels | Inspect form fields | Each input either has `id` + `<label>` or `aria-label` |

---

## Known Issues / Caveats

- **Accessibility warning:** Form fields on `/register` and `/login` trigger "form field element should have an id or name attribute" in console. These are browser-level warnings, not React errors.
- **Private conversations** are created by discovering a user and starting a chat, not automatically when a user is registered.
- **Unread badges** are maintained client-side via WebSocket events. On page refresh, the backend returns the correct count via the conversations API.

---

## Test Environment

- **App URL:** http://localhost:3000
- **API URL:** http://localhost:8080
- **WebSocket:** ws://localhost:8080/ws
- **Test users created during last run:**
  - `alice_e2e_123` / `alice_e2e_123@test.com` / `testpass123`
  - `bob_e2e_123` / `bob_e2e_123@test.com` / `testpass123`

---

## Quick Smoke Test (5 min)

For a quick verification after changes:

1. Open Chrome → DevTools Console
2. Register two users in different windows
3. Verify `Connected` status on both
4. From User A: Discover User B → start conversation → send message
5. Verify User B sees unread badge + message appears in sidebar + chat
6. Verify User B's message is left-aligned, User A's is right-aligned
7. Click conversation → verify unread badge clears
8. Verify console has **no** errors
