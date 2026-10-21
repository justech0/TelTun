<p align="center">
  <img src="build/appicon.png" width="112" alt="TelTun"/>
</p>

<h1 align="center">TelTun</h1>
<p align="center"><b>Share your Android phone's mobile internet with your Windows PC over USB.</b><br/>
Made by Enes · English / Türkçe</p>

---

## English

TelTun is a small Windows program. Plug in your phone with a USB cable, press **Connect**, and your PC uses the phone's mobile data. Nothing to install on the phone.

### Features
- One-click **Connect / Disconnect**
- Live **download / upload speed** with charts, and a **speedtest.net** button
- Public IP is **hidden by default**, with a show / hide button
- **Mobile hotspot** button: share the tunnel with other devices over Wi-Fi
- **Tray icon**, start with Windows, auto reconnect
- **English and Turkish** interface
- A **How it works** page that explains every step — TelTun is made for learning

### First time setup
1. On the phone open **Settings → About phone** and tap **Build number** 7 times. Developer options are now on.
   - Xiaomi / Redmi / POCO: tap **OS version** instead.
   - Samsung: **About phone → Software information → Build number**.
2. Open **Developer options** and turn on **USB debugging**. It is usually under **Settings → System**. On Xiaomi it is under **Additional settings**, on Samsung at the bottom of Settings.
3. Plug in the cable. Unlock the phone and tap **Allow** on the USB debugging prompt (tick *Always allow*).
4. Open `TelTun.exe` (it asks for admin rights) and press **Connect**.

### Tested phones
Works on most Android phones that have USB debugging. Tested and working on:
- Redmi Note 13 Pro 5G
- Redmi Note 10 Pro
- Xiaomi 17 Pro

### How it works
```
PC apps ──► TelTun virtual card ──► USB link (adb) ──► helper on phone ──► mobile network
            (sing-box TUN)          one stream,         opens the real
                                    many connections    connections
```
- TelTun copies a tiny helper to the phone (`/data/local/tmp/teltun`) and starts it with `adb shell`. It listens only on the phone's own loopback (127.0.0.1).
- `adb forward` connects the PC to the helper through USB.
- **Why one stream?** adb carries about 300 Mbps in a single stream, but many adb streams at once stall each other. TelTun packs every connection into **one** stream with [yamux](https://github.com/hashicorp/yamux) multiplexing. Each connection inside is a small SOCKS5 session; UDP travels as *UDP-over-TCP*.
- On the PC, [sing-box](https://github.com/SagerNet/sing-box) creates a virtual network card named **TelTun** and sends all traffic (and DNS) into the link.
- Because the phone opens the real connections, the traffic leaves from the phone like normal phone traffic.
- When you disconnect or quit, the virtual card is removed and your normal network is back.

### Build from source
Needs Go 1.25+, Node.js and the Wails CLI:
```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
./build.ps1          # result: build/bin/TelTun.exe
```
`build.ps1` downloads Google's platform-tools (adb), cross-compiles the phone helper for arm64 / arm / x86_64 and embeds everything into a single exe.

Project layout:
| Path | What |
|---|---|
| `cmd/teltun-phone` | Helper that runs on the phone (SOCKS5 + UDP-over-TCP inside yamux) |
| `internal/engine` | adb control, USB link, sing-box TUN |
| `internal/hotspot` | Windows Mobile hotspot (WinRT via PowerShell) |
| `internal/autostart` | Start with Windows (Task Scheduler) |
| `frontend` | Glass UI (HTML/CSS/JS), translations in `src/i18n.js` |
| `cmd/linktest` | Dev tool: test the phone link without touching routes |

### Notes
- Admin rights are needed to create the virtual network card.
- Logs: `%LOCALAPPDATA%\TelTun\teltun.log` (or the Logs page).

---

## Türkçe

TelTun küçük bir Windows programı. Telefonu USB kabloyla tak, **Bağlan**'a bas, bilgisayar telefonun mobil verisini kullansın. Telefona bir şey kurmana gerek yok.

### Özellikler
- Tek tıkla **Bağlan / Kes**
- Grafikli anlık **indirme / yükleme hızı** ve **speedtest.net** hız testi düğmesi
- Genel IP **varsayılan olarak gizli**, göster / gizle düğmesi var
- **Mobil etkin nokta** düğmesi: tüneli Wi-Fi ile başka cihazlarla paylaş
- **Sistem tepsisi**, Windows ile başlat, kopunca yeniden bağlan
- **Türkçe ve İngilizce** arayüz
- Her adımı anlatan **Nasıl çalışır** sayfası — TelTun öğrenmek için yapıldı

### İlk kurulum
1. Telefonda **Ayarlar → Telefon hakkında** bölümüne gir ve **Yapı numarası**'na 7 kez dokun. Geliştirici seçenekleri açılır.
   - Xiaomi / Redmi / POCO: **OS sürümü**'ne dokun.
   - Samsung: **Telefon hakkında → Yazılım bilgileri → Yapı numarası**.
2. **Geliştirici seçenekleri**'ne gir ve **USB hata ayıklama**'yı aç. Genelde **Ayarlar → Sistem** altındadır. Xiaomi'de **Ek ayarlar** altında, Samsung'da Ayarlar'ın en altındadır.
3. Kabloyu tak. Telefonun kilidini aç, USB hata ayıklama sorusunda **İzin ver**'e dokun (*Her zaman izin ver*'i işaretle).
4. `TelTun.exe`'yi aç (yönetici izni ister) ve **Bağlan**'a bas.

### Test edilen telefonlar
USB hata ayıklaması olan çoğu Android telefonda çalışır. Denenip çalışanlar:
- Redmi Note 13 Pro 5G
- Redmi Note 10 Pro
- Xiaomi 17 Pro

### Nasıl çalışır
- TelTun telefona küçük bir yardımcı kopyalar (`/data/local/tmp/teltun`) ve `adb shell` ile başlatır. Yardımcı sadece telefonun kendi içini (127.0.0.1) dinler.
- `adb forward`, bilgisayarı USB üzerinden yardımcıya bağlar.
- **Neden tek akış?** adb tek akışta yaklaşık 300 Mbps taşır, ama aynı anda çok akış açılınca birbirini kilitler. TelTun bütün bağlantıları [yamux](https://github.com/hashicorp/yamux) ile **tek** akışın içine koyar. İçerdeki her bağlantı küçük bir SOCKS5 oturumudur; UDP, *UDP-over-TCP* olarak taşınır.
- Bilgisayarda [sing-box](https://github.com/SagerNet/sing-box), **TelTun** adlı sanal ağ kartı kurar ve tüm trafiği (DNS dahil) hatta yollar.
- Gerçek bağlantıları telefon açtığı için trafik telefondan normal telefon trafiği gibi çıkar.
- Kesince ya da kapatınca sanal kart silinir, normal internet geri gelir.

### Kaynaktan derleme
Go 1.25+, Node.js ve Wails CLI gerekir. `./build.ps1` çalıştır, sonuç `build/bin/TelTun.exe`.

### Notlar
- Sanal ağ kartı için yönetici izni gerekir.
- Kayıtlar: `%LOCALAPPDATA%\TelTun\teltun.log` (veya Kayıtlar sayfası).

---

## Contact / İletişim
📧 [iletisim@enestunc.com.tr](mailto:iletisim@enestunc.com.tr)

## License / Lisans
GPL-3.0 — see [LICENSE](LICENSE). Uses sing-box (GPL-3.0), yamux (MPL-2.0), Wails (MIT), fyne systray (Apache-2.0) and Android platform-tools (Apache-2.0, see `assets/adb/NOTICE.txt`).

<p align="center"><sub>Made by Enes</sub></p>
