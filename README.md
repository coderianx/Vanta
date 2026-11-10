# Vanta

Vanta, **devasa bir community platformu olma vizyonuyla** Go, PostgreSQL ve JWT kullanılarak geliştirilen bir REST API projesidir. Şu anda kullanıcı kaydı ve girişi için kimlik doğrulama bileşenlerini içerir.

## Vizyon

Vanta'nın amacı; insanların ilgi alanları etrafında bir araya gelebileceği, kendi topluluklarını kurabileceği, içerik paylaşabileceği ve birbirleriyle etkileşime geçebileceği büyük ölçekli bir topluluk platformu oluşturmaktır.

Kullanıcı profilleri, topluluklar, gönderiler, yorumlar, etkileşimler ve moderasyon sistemleri gibi özelliklerle büyüyerek yüksek kullanıcı ve içerik hacmini destekleyen bir platforma dönüşmesi hedeflenmektedir. Mevcut kimlik doğrulama altyapısı, bu vizyonun ilk adımlarından biridir.

> **Durum:** Proje geliştirme aşamasındadır. Kayıt ve giriş handler'ları uygulanmıştır; ancak `cmd/vanta/main.go` içindeki veritabanı başlatma, route kayıtları ve HTTP sunucusu kurulumu henüz tamamlanmamıştır.

## Mevcut Özellikler

- Bcrypt ile güvenli parola hash'leme ve doğrulama
- PostgreSQL üzerinde kullanıcı kaydı ve sorgulama
- JWT ile 24 saat geçerli erişim token'ı üretimi
- SHA-256 özeti veritabanında saklanan, 7 gün geçerli refresh token üretimi
- Chi router ile `/api/v1` altında route organizasyonu
- Handler, repository, model ve veritabanı katmanlarının ayrılması
- Parametreli SQL sorguları ve kullanıcı adı benzersizlik kontrolü

## Teknolojiler

- Go 1.27.1
- PostgreSQL
- Chi
- pgx
- golang-jwt
- bcrypt
- godotenv

## Gereksinimler

- Go 1.27.1 veya uyumlu bir sürüm
- Çalışan bir PostgreSQL veritabanı

## Kurulum

Depoyu klonlayın ve bağımlılıkları indirin:

```bash
git clone https://github.com/coderianx/Vanta.git
cd Vanta
go mod download
```

Proje kökünde bir `.env` dosyası oluşturun:

```env
DB_URL=postgres://postgres:postgres@localhost:5432/vanta?sslmode=disable
JWT_SECRET=guclu-ve-rastgele-bir-secret
```

`JWT_SECRET` değerini gerçek ortamlarda tahmin edilmesi zor, uzun ve rastgele bir değerle değiştirin. `.env` dosyası `.gitignore` tarafından hariç tutulur ve versiyon kontrolüne eklenmemelidir.

## Veritabanı

Uygulama aşağıdaki `users` ve `refresh_tokens` tablolarını kullanır:

```sql
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Tablolar, yukarıdaki SQL ile manuel olarak veya uygulama başlangıcında `store.CreateTables()` çağrılarak oluşturulabilir. Refresh token'ın yalnızca SHA-256 özeti saklanır; ham token yalnızca login yanıtında istemciye gönderilir.

## API

### Kullanıcı kaydı

```http
POST /api/v1/auth/register
Content-Type: application/json
```

```json
{
  "name": "Ada Lovelace",
  "username": "ada",
  "password": "guclu-bir-parola"
}
```

Parola 8 ile 72 bayt arasında olmalıdır. Başarılı kayıt `201 Created` döndürür:

```json
{
  "message": "user registered successfully"
}
```

Kullanıcı adı zaten kayıtlıysa `409 Conflict` döndürülür.

### Kullanıcı girişi

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "username": "ada",
  "password": "guclu-bir-parola"
}
```

Başarılı girişte `200 OK` ile birlikte JWT ve 7 gün geçerli refresh token döndürülür:

```json
{
  "token": "signed-jwt-token",
  "refresh_token": "random-refresh-token"
}
```

Hatalı kullanıcı adı veya parola için `401 Unauthorized` ve `invalid_credentials` yanıtı döndürülür.

## Çalıştırma

Uygulama başlangıç akışı tamamlandığında proje şu komutla çalıştırılabilir:

```bash
go run ./cmd/vanta
```

Mevcut durumda `main.go`, `store.ConnectDatabase()` veya `http.ListenAndServe()` çağırmadığı için komut bir HTTP sunucusu başlatmaz. Login endpoint'inin kullanılabilmesi için ayrıca `/api/v1/auth/login` route'unun `auth.LoginHandler` ile kaydedilmesi gerekir.

## Kontroller

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./...
```

## Proje yapısı

```text
.
├── cmd/vanta/                 # Uygulama giriş noktası
├── internal/
│   ├── handlers/auth/         # Kayıt ve giriş HTTP handler'ları
│   ├── models/                # İstek, kullanıcı ve JWT modelleri
│   ├── repository/            # Kullanıcı veritabanı işlemleri
│   └── store/                 # PostgreSQL bağlantısı ve tablo kurulumu
├── go.mod
├── go.sum
└── README.md
```

## Lisans

Bu proje [MIT Lisansı](LICENSE) ile lisanslanmıştır.
