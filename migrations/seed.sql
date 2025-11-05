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