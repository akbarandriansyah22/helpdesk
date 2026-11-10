# Catatan Arsitektur

## Keputusan Arsitektur

- **Layered architecture** (handler → service → repository) agar logika bisnis terpisah dari HTTP dan database, mudah diuji, dan mudah diganti implementasi.
- **Status transition** divalidasi di service layer, bukan di handler, agar aturan bisnis tidak bisa dilewati.
- **Role-based access** dicek di service dan middleware. Requester hanya bisa mengakses tiket miliknya.
- **JWT** untuk autentikasi stateless. Password di-hash dengan bcrypt.
- **Migrasi SQL idempotent** + seed data agar reviewer bisa menjalankan aplikasi dengan satu perintah.
- **docker-compose** untuk menjalankan backend, frontend, dan database sekaligus.

## Hal yang Sengaja Ditinggalkan

- Upload lampiran (belum ada validasi tipe file dan batas ukuran).
- Notifikasi real-time (WebSocket/SSE).
- Aturan SLA dan halaman statistik admin.
- Refresh token dan logout (JWT masih stateless tanpa blacklist).
- Rate limiting dan audit log.

## Perbaikan jika Ada Waktu Lebih

### Keamanan
- Tambah rate limiting pada endpoint login dan register untuk mencegah brute force.
- Implementasi refresh token + blacklist token saat logout.
- Validasi dan sanitasi input lebih ketat (termasuk XSS pada komentar).
- Tambah audit log untuk perubahan status tiket dan perubahan role.
- Gunakan secret manager atau Docker secrets, bukan environment variable biasa.
- Tambah security headers (CSP, HSTS, dll) di frontend dan backend.
- Lakukan dependency scanning dan SAST di pipeline CI.

### DevOps
- Buat pipeline CI (GitHub Actions) yang menjalankan lint, unit test, dan build Docker image.
- Tambah health check dan readiness probe yang lebih lengkap.
- Pisahkan environment (dev/staging) dengan file compose berbeda.
- Tambah observability dasar (structured logging + metrics sederhana).
- Gunakan multi-stage Dockerfile yang lebih optimal dan non-root user.
- Tambah database migration tool yang lebih formal (misalnya golang-migrate) jika skema berkembang.