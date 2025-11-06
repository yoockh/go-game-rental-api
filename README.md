# Video Game Rental API

## Overview
Video Game Rental API is a backend system built with Golang (Echo Framework) for physical game rental platform including cartridges and consoles.  
This project implements multi-role system (Super Admin, Admin, Partner, Customer), along with payment system, review, and approval flow features.

---

## Tech Stack
| Layer | Technology |
|-------|------------|
| Backend | Go (Echo v4) |
| Database | PostgreSQL / Supabase |
| ORM / Query | GORM |
| Authentication | JWT + Refresh Token |
| File Storage | Supabase Storage |
| Payment Gateway | Stripe / Midtrans |
| Validation | go-playground/validator v10 |
| Logging | logrus |
| Documentation | Swagger (swaggo) |
| CI/CD | Heroku / Railway |
| Testing | testify, mockgen |

---

## Modules & Features

### Implemented

#### Auth
-  Register & Login (default role: `customer`)
-  JWT Authentication with bcrypt password hashing
-  Role-based Access Control (RBAC)
-  Refresh Token (placeholder) -- pending

#### User Management
-  View & Edit Profile
-  Admin user management (view, role update, ban/unban)

#### Partner System
-  Apply for Partner (application form)
-  Admin approve/reject partner applications
-  Partner game listings (CRUD)
-  Admin approve/reject game listings
-  Partner view bookings for their games

#### Game Catalog
-  List Games (public)
-  Game detail view
-  Game search functionality
-  Partner game management

#### Booking System
-  Create booking
-  View user bookings
-  Cancel booking
-  Partner confirm handover/return
-  Admin view all bookings

#### Payment System
-  Basic payment structure
-  Payment webhook handling
-  Payment gateway integration (Stripe/Midtrans) -- pending

#### Review System
-  Create review for completed bookings
-  View game reviews



### In Development
- Refresh token implementation
- Advanced filtering and search
- Analytics dashboard
- Email notifications enhancement

---

## Detailed Business Flow

### Partner Onboarding Flow
1. User register → role: `customer`
2. Customer submits partner application via `/partner/apply` (simple form with business info)
3. Admin reviews application via `/admin/partner-applications/:id/approve`
4. If approved → user.role = `partner`
5. Partner can now create game listings

### Game Listing Flow
1. Partner creates game listing via `/partner/games`
2. Listing status = `pending_approval`
3. Admin reviews listing via `/admin/listings/:id/approve`
4. If approved → game.is_active = true, visible to customers

### Booking & Rental Flow
1. Customer browses approved games via `/games`
2. Customer creates booking via `/bookings` → status: `pending_payment`
3. Customer pays via `/payments/:booking_id/pay`
4. Payment webhook confirms → booking status: `confirmed`
5. **Partner confirms item handover** → status: `active`
6. Customer returns item → Partner confirms return → status: `completed`
7. Customer can leave review via `/bookings/:id/review`

---

## Entity Relationship Diagram (ERD) - Summary
- users (id, email, password, full_name, phone, address, role, is_active, created_at, updated_at)
- categories (id, name, description, is_active, created_at)
- partner_applications (id, user_id, business_name, business_address, business_phone, business_description, status, rejection_reason, submitted_at, decided_at, decided_by)
- games (id, partner_id, category_id, name, description, platform, stock, available_stock, rental_price_per_day, security_deposit, condition, images, is_active, approval_status, approved_by, approved_at, rejection_reason, created_at, updated_at)
- bookings (id, user_id, game_id, partner_id, start_date, end_date, rental_days, daily_price, total_rental_price, security_deposit, total_amount, status, notes, handover_confirmed_at, return_confirmed_at, created_at, updated_at)
- payments (id, booking_id, provider, provider_payment_id, amount, status, payment_method, paid_at, failed_at, failure_reason, created_at)
- reviews (id, booking_id, user_id, game_id, rating, comment, created_at, updated_at)
- refresh_tokens (id, user_id, token_hash, expires_at, is_revoked, created_at)

---

## API Endpoint Pattern

| Resource | Method | Endpoint | Description |
|-----------|---------|----------|-------------|
| **Auth** | POST | /auth/register | Register user |
|  | POST | /auth/login | Login |
|  | POST | /auth/refresh | Refresh token |
| **Users** | GET | /users/me | Get current user profile |
|  | PUT | /users/me | Update profile |
| **Partner** | POST | /partner/apply | Submit partner application |
|  | GET | /partner/applications | Get all partner applications *(admin only)* |
|  | PATCH | /partner/applications/:id/approve | Approve or reject partner *(admin)* |
| **Games** | GET | /games | Get all games |
|  | GET | /games/:id | Get game detail |
|  | POST | /partner/games | Create new game listing *(partner)* |
|  | PATCH | /partner/games/:id | Update own game listing *(partner)* |
| **Bookings** | POST | /bookings | Create booking *(customer)* |
|  | GET | /bookings/:id | Get booking detail *(authorized only)* |
|  | PATCH | /bookings/:id/cancel | Cancel booking *(customer)* |
| **Payments** | POST | /payments/:booking_id/pay | Make payment |
|  | POST | /webhooks/payments | Handle payment webhook *(system)* |
| **Reviews** | POST | /bookings/:id/review | Add review after completed booking |
| **Admin** | GET | /admin/users | View all users |
|  | PATCH | /admin/users/:id/ban | Ban / unban user |
|  | GET | /admin/partner-applications | View pending partner applications |
|  | PATCH | /admin/partner-applications/:id/approve | Approve / reject partner application |
|  | GET | /admin/listings | View pending listings |
|  | PATCH | /admin/listings/:id/approve | Approve / reject listing |
| **Superadmin** | GET | /superadmin/admins | View all admins |
|  | POST | /superadmin/admins | Create new admin |
|  | DELETE | /superadmin/admins/:id | Remove admin |

| **Partner Dashboard** | GET | /partner/dashboard | Partner analytics |
|  | GET | /partner/bookings | View bookings for partner's games |
|  | PATCH | /partner/bookings/:id/confirm-handover | Confirm item handover |
|  | PATCH | /partner/bookings/:id/confirm-return | Confirm item return |

---

## Security & Authentication

### Public Endpoints (No authentication required)
- `POST /auth/register` - User registration
- `POST /auth/login` - User login  
- `POST /auth/refresh` - Refresh JWT token
- `GET /games` - Browse game catalog
- `GET /games/:id` - View game details
- `POST /webhooks/payments` - Payment gateway webhooks

### Protected Endpoints
All other endpoints require valid JWT token in Authorization header:
```
Authorization: Bearer <jwt_token>
```

### Role-Based Access Control (RBAC)
- **Customer**: Can book games, manage profile
- **Partner**: Customer permissions + manage game listings, view bookings
- **Admin**: Partner permissions + approve applications/listings, handle disputes
- **Super Admin**: Full system access + manage admins

---

## Status Definitions

### User Roles
- `customer` - Default role, can book games
- `partner` - Can list games for rental
- `admin` - Can approve/reject applications and listings
- `super_admin` - Full system access

### Booking Status
- `pending_payment` - Awaiting payment
- `confirmed` - Payment received
- `active` - Item handed over, rental in progress
- `completed` - Item returned successfully
- `cancelled` - Booking cancelled

### Payment Status
- `pending` - Payment initiated
- `paid` - Payment successful
- `failed` - Payment failed
- `refunded` - Payment refunded

### Application Status
- `pending` - Partner application submitted, awaiting review
- `approved` - Application approved
- `rejected` - Application rejected

---

## Third-Party Integration

###  Implemented
- **Database**: PostgreSQL with GORM
- **Documentation**: Swagger (swaggo) - auto-generated
- **Validation**: go-playground/validator v10
- **Logging**: logrus
- **Storage**: Supabase Storage (for game images)
- **Payment Gateway**: Midtrans (sandbox mode)
- **Email Notification**: SendGrid

###  Planned
- **Error Tracking**: Sentry
- **Deployment**: Heroku / Railway
- **Advanced Analytics**: Custom dashboard

---

## Setup Guide
1. Clone repository
   ```bash
   git clone https://github.com/Yoochan45/go-game-rental-api.git
   cd go-game-rental-api
   ```

2. Install dependencies
   ```bash
   go mod tidy
   ```

3. Setup environment variables
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run the application
   ```bash
   go run app/echo-server/main.go
   ```

5. Access Swagger documentation
   ```
   http://localhost:8080/swagger/index.html
   ```

### Development Status
-  **Core API**: Fully functional with all basic CRUD operations
-  **Authentication**: JWT-based auth with role-based access control
-  **Database**: PostgreSQL with GORM, auto-migration
-  **Documentation**: Complete Swagger API docs
-  **Clean Architecture**: Handler → Service → Repository pattern
-  **3rd Party**: Payment gateways, file storage, email (planned)

---
## Contributor
Aisiya Qutwatunnada (Yoochan45)