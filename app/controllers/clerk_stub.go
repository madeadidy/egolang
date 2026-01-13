package controllers

/*
  clerk_stub.go

  File ini berisi penjelasan dan stub untuk integrasi Clerk server-side.
  - Untuk integrasi penuh, install Clerk Go SDK sesuai dokumentasi resmi, misal:
      go get github.com/clerkinc/clerk-sdk-go

  - Setelah SDK terinstal, implementasikan inisialisasi client dan middleware
    yang memverifikasi session/token dari header/cookie, lalu ambil data user
    dari Clerk dan set di context request.

  Contoh (pseudocode):

    func NewClerkClient() *clerk.Client {
        secret := os.Getenv("CLERK_SECRET_KEY")
        return clerk.NewClient(clerk.WithSecretKey(secret))
    }

    func (server *Server) ClerkAuthMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // baca token dari cookie atau header
            // verifikasi menggunakan client
            // jika valid, set user info ke context dan lanjut
            next.ServeHTTP(w, r)
        })
    }

  Catatan:
  - Jangan commit `CLERK_SECRET_KEY` ke repo.
  - Sesuaikan bagaimana frontend mengirim session token (cookie vs Authorization header).
*/
