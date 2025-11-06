-- Seeder for users (4 roles)
INSERT INTO users (email, password_hash, full_name, phone, address, role)
VALUES
-- Super Admin
('superadmin@example.com', '$2a$06$Zy0/q5.34gI.aWqIvJO6y./BlYHDuMXEKncPIXJDvARA77gfdRL1i', 'Super Admin', '081234567890', 'HQ Building, Jakarta', 'super_admin'),

-- Admin
('admin@example.com', '$2a$06$wP/18z9pBgUbAM3P35S/s.IeNBRjOnl/bEqSOVSt2rUwP4zcLyGtG', 'GameZone Admin', '081234567891', 'Jl. Admin No. 1, Bandung', 'admin'),

-- Partner
('partner@example.com', '$2a$06$ryLe.YwU/83QmqLh0hp7CuH83YdGv7Dr/PXBNaZ29mT.dEh6S6RiO', 'Game Partner', '081234567892', 'Jl. Mitra No. 2, Surabaya', 'partner'),

-- Customer
('customer@example.com', '$2a$06$EMWsaivLirtHGeyAjOo.UOwDglis7krsH44V.NFAwzl1hjhwvghbi', 'Customer User', '081234567893', 'Jl. Pelanggan No. 3, Yogyakarta', 'customer');

-- Default categories
INSERT INTO categories (name, description) VALUES 
('Action', 'Action and adventure games'),
('RPG', 'Role-playing games'),
('Strategy', 'Strategy and simulation games'),
('Sports', 'Sports games'),
('Racing', 'Racing games'),
('Fighting', 'Fighting games'),
('Puzzle', 'Puzzle and brain games'),
('Horror', 'Horror and thriller games');

-- Insert Games (partner_id=2 is partner1@example.com)
INSERT INTO games (partner_id, category_id, name, description, platform, stock, available_stock, rental_price_per_day, security_deposit, condition, is_active, approval_status, approved_by, approved_at) VALUES
(2, 1, 'God of War Ragnarök', 'Epic Norse mythology adventure game', 'PlayStation 5', 3, 3, 15000.00, 50000.00, 'excellent', true, 'approved', 1, NOW()),
(2, 1, 'Spider-Man: Miles Morales', 'Superhero action-adventure game', 'PlayStation 5', 2, 2, 12000.00, 45000.00, 'excellent', true, 'approved', 1, NOW()),
(2, 2, 'Halo Infinite', 'Sci-fi first-person shooter', 'Xbox Series X', 2, 2, 13000.00, 40000.00, 'good', true, 'approved', 1, NOW()),
(2, 3, 'The Legend of Zelda: Breath of the Wild', 'Open-world adventure game', 'Nintendo Switch', 4, 4, 10000.00, 35000.00, 'excellent', true, 'approved', 1, NOW()),
(2, 3, 'Super Mario Odyssey', 'Platform adventure game', 'Nintendo Switch', 3, 3, 8000.00, 30000.00, 'good', true, 'approved', 1, NOW());

-- Insert Sample Bookings (user_id=3 is customer1@example.com, game_id=1 is God of War)
INSERT INTO bookings (user_id, game_id, partner_id, start_date, end_date, rental_days, daily_price, total_rental_price, security_deposit, total_amount, status) VALUES
(3, 1, 2, '2024-01-15', '2024-01-17', 3, 15000.00, 45000.00, 50000.00, 95000.00, 'completed'),
(4, 2, 2, '2024-01-20', '2024-01-22', 3, 12000.00, 36000.00, 45000.00, 81000.00, 'active');

-- Insert Sample Payments
INSERT INTO payments (booking_id, provider, amount, status, payment_method, paid_at) VALUES
(1, 'midtrans', 95000.00, 'paid', 'bank_transfer', '2024-01-15 10:30:00'),
(2, 'midtrans', 81000.00, 'paid', 'credit_card', '2024-01-20 14:15:00');

-- Insert Sample Reviews
INSERT INTO reviews (booking_id, user_id, game_id, rating, comment) VALUES
(1, 3, 1, 5, 'Amazing game! Graphics are stunning and gameplay is smooth. Highly recommended!');