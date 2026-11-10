Berikut isi lengkap **README.md** yang bisa langsung kamu copy-paste ke root repo (`helpdesk/README.md`):

```markdown
# Helpdesk Internal

Aplikasi helpdesk sederhana untuk mengelola tiket dukungan karyawan. Dibangun dengan REST API Go dan antarmuka Vue 3.

## Fitur

- Autentikasi JWT dengan tiga peran: Requester, Agent, Admin
- CRUD tiket dengan aturan transisi status
- Komentar publik dan catatan internal
- Daftar tiket dengan pagination, filter, pencarian, dan sorting
- Manajemen pengguna (admin)

## Teknologi

| Layer     | Teknologi                            |
| --------- | ------------------------------------ |
| Backend   | Go 1.22+, Gin, PostgreSQL            |
| Frontend  | Vue 3 (Composition API), Vite, Pinia |
| Database  | PostgreSQL 16                        |
| Container | Docker Compose                       |

## Cara Menjalankan

### Prasyarat

- Docker dan Docker Compose

### Langkah

1. Clone repository
```bash
git clone https://github.com/akbarandriansyah22/helpdesk.git
cd helpdesk
```

2. Jalankan semua service
```bash
cd backend
docker compose up --build
```

3. Akses aplikasi
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

Migrasi database dan seed data berjalan otomatis saat backend startup.

## Akun Seed

Password semua akun: `password123`

| Peran     | Email                 | 
| --------- | --------------------- |
| Requester | requester@example.com |
| Agent     | agent@example.com     |
| Admin     | admin@example.com     | 

## Daftar Endpoint

| Method | Path                       | Deskripsi                             | Akses               |
| ------ | -------------------------- | ------------------------------------- | ------------------- |
| POST   | /api/auth/register         | Daftar akun requester                 | Publik              |
| POST   | /api/auth/login            | Login, mengembalikan JWT              | Publik              |
| GET    | /api/me                    | Profil pengguna saat ini              | Login               |
| GET    | /api/tickets               | Daftar tiket (filter, pagination)     | Login               |
| POST   | /api/tickets               | Buat tiket                            | Login               |
| GET    | /api/tickets/{id}          | Detail tiket beserta komentar         | Pemilik/Agent/Admin |
| PATCH  | /api/tickets/{id}          | Ubah status, prioritas, assignee      | Sesuai peran        |
| POST   | /api/tickets/{id}/comments | Tambah komentar atau catatan internal | Pemilik/Agent/Admin |
| GET    | /api/users                 | Daftar pengguna                       | Admin/Agent         |
| PATCH  | /api/users/{id}/role       | Ubah peran pengguna                   | Admin               |

Query parameter untuk `GET /api/tickets`:
- `page`, `limit` (default 1 dan 20, maksimal 100)
- `status`, `priority`, `assignee_id` (`unassigned` untuk tiket tanpa assignee)
- `search` (cari judul)
- `sort_by` (`created_at`, `updated_at`, `priority`, `status`, `title`)
- `order` (`asc` atau `desc`)

## Migrasi Database

Migrasi berjalan otomatis saat API startup. Untuk menjalankan manual:

```bash
cd backend
cp .env.example .env
# Isi DB_PASSWORD dan JWT_SECRET
go run ./cmd/migrate
```

File migrasi:
- `backend/migrations/001_init.sql`
- `backend/migrations/002_seed.sql`

Kedua file aman dijalankan berulang kali.

## Menjalankan Test

```bash
cd backend
go test ./...
```

## Dokumentasi API

Koleksi Thunder Client tersedia di `docs/helpdesk.thunder-collection.json`.

## Struktur Proyek

```
helpdesk/
├── backend/          # REST API (Go)
├── frontend/         # Antarmuka (Vue 3)
├── docs/             # Dokumentasi API
└── README.md
```

## Catatan

- Konfigurasi backend melalui environment variable (`backend/.env.example`)
- Transisi status tiket divalidasi di service layer
- Requester hanya dapat melihat dan mengomentari tiket miliknya
- Catatan internal hanya terlihat oleh agent dan admin

