# Video Game Rental API

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Echo](https://img.shields.io/badge/Echo-v4-00ADD8?style=flat)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?style=flat&logo=postgresql)
![License](https://img.shields.io/badge/license-Educational-green)
![Test Coverage](https://img.shields.io/badge/coverage-78.5%25-brightgreen)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)


- **Heroku Link** : https://go-game-rental-3beef3913ef8.herokuapp.com
- **Swagger Link** : https://go-game-rental-3beef3913ef8.herokuapp.com/swagger/index.html

## Overview
Video Game Rental API is a backend system built with Golang (Echo Framework) for physical game rental platform including cartridges and consoles.  
This project implements multi-role system (Super Admin, Admin, Customer), along with payment system, review, and booking flow features.

---

## Tech Stack
| Layer | Technology |
|-------|------------|
| Backend | Go (Echo v4) |
| Database | PostgreSQL / Supabase |
| ORM / Query | GORM |
| Authentication | JWT |
| File Storage | Supabase Storage |
| Payment Gateway | Midtrans |
| Email Service | SendGrid |
| Validation | go-playground/validator v10 |
| Logging | logrus |
| Documentation | Swagger (swaggo) |
| Testing | testify, mock |

---

## Modules & Features

### **Implemented**

#### Auth & User Management
- Register & Login (default role: `customer`)
- JWT Authentication with bcrypt password hashing
- Role-based Access Control (RBAC): `super_admin`, `admin`, `customer`
- View & Edit Profile
- Admin user management (view, role update, activate/deactivate)

#### Game Catalog
- List Games (public) with pagination
- Game detail view
- Game search functionality
- Admin game management (CRUD)
- Category management (CRUD)

#### Booking System
- Create booking
- View user bookings
- View booking detail
- Cancel booking
- Admin view all bookings
- Admin update booking status (confirm/active/complete)

#### Payment System
- Create payment for booking
- Payment webhook handling
- View payment by booking
- Admin view all payments
- Midtrans integration structure

#### Review System
- Create review for completed bookings
- View game reviews (public)

---

### **In Development / Planned**
- Refresh token implementation
- Email notification triggers (welcome, booking confirmation, etc.)
- Advanced filtering (by category, platform, price range)
- Admin analytics dashboard
- Payment gateway full integration (Midtrans/Stripe)

---

## Detailed Business Flow

### User Journey
1. **Register** - Default role: `customer`
2. **Browse Games** - Public access, view catalog
3. **Create Booking** - Select game, dates, Status: `pending_payment`
4. **Make Payment** - Via Midtrans, Status: `confirmed` (after webhook)
5. **Admin Confirms Handover** - Status: `active`
6. **Return Game** - Admin confirms return, Status: `completed`
7. **Leave Review** - Customer rates and reviews

### Admin Workflow
1. **Manage Games** - Add/Edit/Delete games
2. **Manage Categories** - Organize game catalog
3. **Manage Bookings** - View all bookings, update status
4. **Manage Payments** - Track payment history
5. **Manage Users** - View users, change roles, activate/deactivate

---

## Entity Relationship Diagram (ERD) - Summary

```
users
├── id (PK)
├── email (unique)
├── password (bcrypt hashed)
├── full_name
├── phone
├── address
├── role (super_admin, admin, customer)
├── is_active (boolean)
└── timestamps

categories
├── id (PK)
├── name
├── description
├── is_active
└── created_at

games
├── id (PK)
├── admin_id (FK → users)
├── category_id (FK → categories)
├── name
├── description
├── platform
├── stock
├── available_stock
├── rental_price_per_day
├── security_deposit
├── condition (excellent, good, fair)
├── images (text[])
├── is_active
└── timestamps

bookings
├── id (PK)
├── user_id (FK → users)
├── game_id (FK → games)
├── start_date
├── end_date
├── rental_days (calculated)
├── daily_price
├── total_rental_price
├── security_deposit
├── total_amount
├── status (pending_payment, confirmed, active, completed, cancelled)
├── notes
└── timestamps

payments
├── id (PK)
├── booking_id (FK → bookings)
├── provider (midtrans/stripe)
├── provider_payment_id
├── amount
├── status (pending, paid, failed, refunded)
├── payment_method
├── paid_at
├── failed_at
├── failure_reason
└── created_at

reviews
├── id (PK)
├── booking_id (FK → bookings)
├── user_id (FK → users)
├── game_id (FK → games)
├── rating (1-5)
├── comment
└── timestamps
```

---

## API Endpoint Pattern

### Public Endpoints (No Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/register | Register new user |
| POST | /auth/login | Login user |
| GET | /games | Get all games (paginated) |
| GET | /games/:id | Get game detail |
| GET | /games/search?q=query | Search games |
| GET | /categories | Get all categories |
| GET | /categories/:id | Get category detail |
| GET | /games/:id/reviews | Get game reviews |

### Customer Endpoints (Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /users/me | Get current user profile |
| PUT | /users/me | Update profile |
| POST | /bookings | Create new booking |
| GET | /bookings/my | Get my bookings |
| GET | /bookings/:id | Get booking detail |
| PATCH | /bookings/:id/cancel | Cancel booking |
| POST | /bookings/:id/payments | Create payment for booking |
| GET | /bookings/:id/payments | Get payment by booking |
| POST | /bookings/:id/reviews | Create review (after completed) |

### Admin Endpoints (Admin/Super Admin Only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /admin/users | Get all users |
| GET | /admin/users/:id | Get user detail |
| PATCH | /admin/users/:id/role | Update user role |
| PATCH | /admin/users/:id/status | Activate/deactivate user |
| POST | /admin/games | Create game |
| PUT | /admin/games/:id | Update game |
| DELETE | /admin/games/:id | Delete game |
| POST | /admin/categories | Create category |
| PUT | /admin/categories/:id | Update category |
| DELETE | /admin/categories/:id | Delete category |
| GET | /admin/bookings | Get all bookings |
| PATCH | /admin/bookings/:id/status | Update booking status |
| GET | /admin/payments | Get all payments |
| GET | /admin/payments/:id | Get payment detail |
| GET | /admin/payments/status?status=pending | Get payments by status |

### Super Admin Only
| Method | Endpoint | Description |
|--------|----------|-------------|
| DELETE | /admin/users/:id | Delete user permanently |

### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /webhooks/payments | Payment provider callback |

---

## Security & Authentication

### Authentication
- JWT-based authentication
- Token expiry: 24 hours
- Password hashing: bcrypt (cost: 10)

### Authorization Header
```
Authorization: Bearer <jwt_token>
```

### Role-Based Access Control (RBAC)
| Role | Permissions |
|------|-------------|
| **customer** | Browse games, create bookings, manage own profile, leave reviews |
| **admin** | Customer permissions + manage games/categories, view all bookings/payments, manage users (except delete) |
| **super_admin** | Full system access including user deletion |

### Role Management Security Rules
| Action | Customer | Admin | Super Admin |
|--------|----------|-------|-------------|
| View Users | No | Yes | Yes |
| Update Role to Customer | No | Yes | Yes |
| Update Role to Admin | No | Yes | Yes |
| Update Role to Super Admin | No | **No** | Yes |
| Modify Super Admin Role | No | **No** | Yes |
| Toggle Customer Status | No | Yes | Yes |
| Toggle Admin Status | No | Yes | Yes |
| Toggle Super Admin Status | No | **No** | Yes |
| Delete Customer | No | Yes | Yes |
| Delete Admin | No | Yes | Yes |
| Delete Super Admin | No | **No** | Yes |
| Delete Self | No | **No** | **No** |

**Key Security Points:**
- Admin **cannot** promote users to Super Admin
- Admin **cannot** modify Super Admin role or status
- Admin **cannot** delete Super Admin
- **Nobody** can delete themselves
- Only Super Admin can manage Super Admin accounts

---

## Status Definitions

### User Roles
- `customer` - Default role, can book games
- `admin` - Can manage catalog and bookings
- `super_admin` - Full system access

### Booking Status Flow
```
pending_payment → confirmed → active → completed
                      ↓
                  cancelled
```

### Payment Status
- `pending` - Payment initiated
- `paid` - Payment successful
- `failed` - Payment failed
- `refunded` - Payment refunded

---

## Database Configuration

### Supabase Pooler Fix
Due to Supabase connection pooler limitations with prepared statements:

```go
// Disable prepared statements globally
db, err := gorm.Open(postgres.New(postgres.Config{
    DSN:                  dbURL,
    PreferSimpleProtocol: true,
}), &gorm.Config{
    PrepareStmt: false,
})
```

### Connection Pool Settings
```go
sqlDB.SetMaxOpenConns(1)
sqlDB.SetMaxIdleConns(0)
sqlDB.SetConnMaxLifetime(500 * time.Millisecond)
```

---

## Third-Party Integration

### **Implemented**
- **Database**: PostgreSQL (Supabase)
- **ORM**: GORM with auto-migration
- **Documentation**: Swagger (swaggo)
- **Validation**: go-playground/validator v10
- **Logging**: logrus
- **Authentication**: JWT with bcrypt
- **Email**: SendGrid (configured, pending full implementation)
- **Payment**: Midtrans structure (webhook handler ready)

### **Planned**
- **Error Tracking**: Sentry
- **Deployment**: Railway / Heroku
- **Monitoring**: Prometheus + Grafana

---

## Setup Guide

### Prerequisites
- Go 1.21+
- PostgreSQL 14+ (or Supabase account)
- Midtrans sandbox account (optional, for payment testing)
- SendGrid API key (optional, for email)

### Installation

1. **Clone repository**
   ```bash
   git clone https://github.com/Yoochan45/go-game-rental-api.git
   cd go-game-rental-api
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Setup environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

   Required variables:
   ```env
   DATABASE_URL=postgresql://user:password@host:5432/dbname
   JWT_SECRET=your-secret-key
   SENDGRID_API_KEY=your-sendgrid-key
   MIDTRANS_SERVER_KEY=your-midtrans-key
   ```

4. **Run database migrations**
   ```bash
   psql "$DATABASE_URL" -f migrations/ddl.sql
   psql "$DATABASE_URL" -f migrations/seed.sql
   ```

5. **Generate Swagger docs**
   ```bash
   swag init -g app/echo-server/main.go -o ./docs
   ```

6. **Run the application**
   ```bash
   go run app/echo-server/main.go
   ```

7. **Access API**
   - API: `http://localhost:8080`
   - Swagger: `http://localhost:8080/swagger/index.html`

---

## Testing

![Test Status](https://img.shields.io/badge/tests-6%20passed-brightgreen)
![Coverage](https://img.shields.io/badge/coverage-78.5%25-brightgreen)

### Run Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./internal/handler -v -cover

# Generate coverage report
go test ./internal/handler -coverprofile=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open coverage report in browser
xdg-open coverage.html  # Linux
open coverage.html      # macOS
```

### Test Coverage Summary
- **Auth Handler**: 78.5% coverage
- **6 Test Cases**: All passing
  - Register Success
  - Register Validation Error
  - Register Email Exists
  - Login Success
  - Login Invalid Credentials
  - Login Validation Error

### Sample Credentials (from seed data)
```
Super Admin:
- Email: superadmin@example.com
- Password: admin123

Admin:
- Email: admin@example.com
- Password: admin123

Customer:
- Email: customer@example.com
- Password: customer123
```

### Example API Calls

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password123","full_name":"Test User","phone":"081234567890","address":"Jakarta"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"customer@example.com","password":"customer123"}'

# Get games
curl http://localhost:8080/games

# Create booking (requires token)
curl -X POST http://localhost:8080/bookings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"game_id":1,"start_date":"2025-11-10","end_date":"2025-11-15","notes":"Weekend rental"}'
```

---

## Project Structure

```
go-game-rental-api/
├── app/
│   └── echo-server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── dto/                     # Data Transfer Objects
│   ├── handler/                 # HTTP handlers
│   │   ├── auth_handler.go
│   │   └── auth_handler_test.go # Unit tests
│   ├── middleware/              # Custom middleware
│   ├── model/                   # Database models
│   ├── repository/              # Data access layer
│   ├── routes/                  # Route definitions
│   ├── service/                 # Business logic layer
│   └── utils/                   # Helper functions
├── migrations/
│   ├── ddl.sql                  # Database schema
│   └── seed.sql                 # Initial data
├── docs/                        # Swagger documentation
├── coverage.html                # Test coverage report
├── go.mod
├── go.sum
└── README.md
```

---

## Development Status
- **Core API**: Fully functional
- **Authentication**: JWT-based with RBAC
- **Database**: PostgreSQL with GORM, Supabase pooler fix applied
- **Clean Architecture**: Handler → Service → Repository
- **Documentation**: Complete Swagger docs
- **Testing**: Unit tests with mocks, 78.5% coverage
- **3rd Party**: Email/Payment/Storage (structure ready, pending full integration)

---

## Known Issues & Solutions

### Supabase Pooler Prepared Statement Error
**Problem**: `ERROR: prepared statement already exists`  
**Solution**: Disabled prepared statements globally (see Database Configuration)

### Date Format in Booking
**Problem**: Frontend sends `YYYY-MM-DD`, backend expects RFC3339  
**Solution**: Parse date strings manually in handler

### Test Error Logs
**Note**: Error logs during tests are **intentional** for testing error scenarios:
```
time="..." level=error msg="Registration failed" error="email already exists"
time="..." level=error msg="Login failed" email=... error="invalid credentials"
```
These logs appear during `TestRegister_EmailExists` and `TestLogin_InvalidCredentials` test cases.

---

## Contributing
This is a classroom project. For collaboration:
1. Fork the repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

---

## License
This project is created for educational purposes.

---

## Contributor
**Aisiya Qutwatunnada (Yoochan45)**  
GitHub: [@Yoochan45](https://github.com/Yoochan45)

---

## Quick Links
- [Swagger Documentation](http://localhost:8080/swagger/index.html)
- [Coverage Report](./coverage.html)
- [GitHub Repository](https://github.com/Yoochan45/go-game-rental-api)