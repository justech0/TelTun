// All UI text. Keys used by the Go side (steps, errors) live here too.
export const dict = {
  en: {
    navHome: 'Home', navHotspot: 'Hotspot', navGuide: 'How it works', navLogs: 'Logs',
    navSettings: 'Settings', navAbout: 'About',
    madeBy: 'Made by Enes',

    stIdle: 'Not connected', stConnecting: 'Connecting…', stConnected: 'Connected',
    stDisconnecting: 'Disconnecting…',
    heroIdle: 'Phone internet on your PC',
    subIdle: 'Plug in your phone with USB and press the button.',
    subConnected: 'Your PC is using the phone\'s internet.',
    btnConnect: 'Connect', btnDisconnect: 'Disconnect', btnWait: 'Please wait',

    stepPrepare: 'Preparing tools', stepDevice: 'Looking for the phone', stepPush: 'Sending helper to the phone',
    stepPhone: 'Starting helper on the phone', stepLink: 'Opening the USB link', stepTest: 'Testing the internet',
    stepTun: 'Creating the virtual network card',

    errNoDevice: 'No phone found. Plug it in and turn on USB debugging.',
    errUnauthorized: 'The phone is asking for permission. Unlock it and tap "Allow USB debugging".',
    errOffline: 'The phone is offline. Unplug it and plug it in again.',
    errAdb: 'adb could not start.', errPrepare: 'Could not unpack the tools.',
    errPush: 'Could not send the helper to the phone.', errPhone: 'The helper did not start on the phone.',
    errForward: 'Could not open the USB port forward.', errLink: 'Could not open the USB link.',
    errTest: 'The phone has no internet. Check mobile data.', errTun: 'Could not create the virtual network card.',
    errUnknown: 'Something went wrong.', lost: 'The link was lost. Your normal internet is back.',

    download: 'Download', upload: 'Upload', device: 'Phone', publicIp: 'Public IP', uptime: 'Uptime',
    totalData: 'Data used', connections: 'Connections', none: '—', anyPhone: 'Any phone',
    phonesFound: 'Phones', noPhones: 'No phone connected',
    stUnauthorized: 'waiting for permission', stOffline: 'offline', stDevice: 'ready',

    hsTitle: 'Mobile hotspot', hsDesc: 'Share the tunnel with other devices over Wi-Fi, using Windows Mobile hotspot.',
    hsOn: 'Hotspot is on', hsOff: 'Hotspot is off', hsBusy: 'Working…',
    hsStart: 'Turn on', hsStop: 'Turn off', hsName: 'Network name', hsPass: 'Password',
    hsClients: 'Devices joined', hsCustom: 'Custom name and password', hsCustomDesc: 'Leave empty to keep the Windows settings.',
    hsWithTun: 'Turn on with the tunnel', hsWithTunDesc: 'Start the hotspot each time the tunnel connects.',
    hsTip: 'Connect the tunnel first. The hotspot then shares the phone\'s internet.',
    copy: 'Copy', copied: 'Copied', show: 'Show', hide: 'Hide', save: 'Save', saved: 'Saved',

    gTitle: 'How TelTun works', gLead: 'TelTun gives your PC the phone\'s mobile internet through the USB cable. Here is what happens, step by step.',
    gSetupTitle: 'First time setup (one time only)',
    g1t: 'Turn on Developer options', g1d: 'Settings → About phone → tap "Build number" (or "MIUI / OS version") 7 times.',
    g2t: 'Turn on USB debugging', g2d: 'Settings → Additional settings → Developer options → USB debugging.',
    g3t: 'Plug in the cable', g3d: 'Use a good data cable. Pick "File transfer" if the phone asks.',
    g4t: 'Allow this PC', g4d: 'Unlock the phone and tap "Allow" on the USB debugging prompt. Tick "Always allow".',
    gFlowTitle: 'The path of your data',
    fApps: 'PC apps', fAppsD: 'Browser, games, updates',
    fTun: 'TelTun card', fTunD: 'Virtual network card takes all traffic',
    fUsb: 'USB link', fUsbD: 'One adb stream, many connections inside',
    fPhone: 'Phone helper', fPhoneD: 'Opens the real connections',
    fNet: 'Mobile network', fNetD: '4G / 5G internet',
    gWhyTitle: 'Why one stream?',
    gWhy: 'adb can carry about 300 Mbps in a single stream, but many streams at once slow each other down. So TelTun packs every connection into one stream (multiplexing). That is why it is fast.',
    gAdminTitle: 'Why admin rights?',
    gAdmin: 'Creating a virtual network card and changing routes needs administrator rights. TelTun asks once when it opens.',
    gSafeTitle: 'Is it safe?',
    gSafe: 'The helper on the phone listens only on the phone itself (127.0.0.1). Nothing is open to the network. When you disconnect or quit, the virtual card is removed and your normal internet comes back.',
    gTroubleTitle: 'If something goes wrong',
    gTrouble: 'Disconnect, unplug the cable, plug it back in and connect again. Check the Logs page for details.',

    logsTitle: 'Logs', logsOpen: 'Open folder', logsCopy: 'Copy all', logsEmpty: 'Nothing yet.',

    setTitle: 'Settings', setLang: 'Language', setLangDesc: 'Interface language.',
    setAutostart: 'Start with Windows', setAutostartDesc: 'Open in the tray when you sign in.',
    setAutoConnect: 'Connect when TelTun opens', setAutoConnectDesc: 'Connect as soon as a phone is found.',
    setAutoReconnect: 'Reconnect after a drop', setAutoReconnectDesc: 'If the cable comes loose, connect again when the phone is back.',
    setTray: 'Close to tray', setTrayDesc: 'The X button hides the window. Quit from the tray menu.',
    setQuic: 'Block QUIC (UDP 443)', setQuicDesc: 'Faster for most sites. Browsers switch to normal HTTPS.',
    setDevice: 'Phone to use', setDeviceDesc: 'Pick one if you plug in more than one phone.',
    setApplyNext: 'Takes effect on the next connect.',

    abTitle: 'About', abDesc: 'Share your phone\'s mobile data with your Windows PC over USB. Free and open source, made for learning.',
    abMadeBy: 'Made by Enes', abVersion: 'Version', abLicense: 'License', abSource: 'Source code',
    abCredits: 'Built with', abNote: 'Use it with your own phone and mobile plan, and follow your operator\'s terms.',
  },
  tr: {
    navHome: 'Ana sayfa', navHotspot: 'Etkin nokta', navGuide: 'Nasıl çalışır', navLogs: 'Kayıtlar',
    navSettings: 'Ayarlar', navAbout: 'Hakkında',
    madeBy: 'Made by Enes',

    stIdle: 'Bağlı değil', stConnecting: 'Bağlanıyor…', stConnected: 'Bağlı',
    stDisconnecting: 'Kesiliyor…',
    heroIdle: 'Telefon interneti bilgisayarında',
    subIdle: 'Telefonu USB ile tak ve düğmeye bas.',
    subConnected: 'Bilgisayar telefonun internetini kullanıyor.',
    btnConnect: 'Bağlan', btnDisconnect: 'Kes', btnWait: 'Bekle',

    stepPrepare: 'Araçlar hazırlanıyor', stepDevice: 'Telefon aranıyor', stepPush: 'Yardımcı telefona gönderiliyor',
    stepPhone: 'Telefonda yardımcı başlıyor', stepLink: 'USB hattı açılıyor', stepTest: 'İnternet deneniyor',
    stepTun: 'Sanal ağ kartı kuruluyor',

    errNoDevice: 'Telefon bulunamadı. Tak ve USB hata ayıklamayı aç.',
    errUnauthorized: 'Telefon izin istiyor. Kilidi aç ve "USB hata ayıklamaya izin ver"e dokun.',
    errOffline: 'Telefon çevrim dışı. Çıkarıp yeniden tak.',
    errAdb: 'adb başlatılamadı.', errPrepare: 'Araçlar açılamadı.',
    errPush: 'Yardımcı telefona gönderilemedi.', errPhone: 'Yardımcı telefonda başlamadı.',
    errForward: 'USB port yönlendirme açılamadı.', errLink: 'USB hattı açılamadı.',
    errTest: 'Telefonda internet yok. Mobil veriyi kontrol et.', errTun: 'Sanal ağ kartı kurulamadı.',
    errUnknown: 'Bir şeyler ters gitti.', lost: 'Bağlantı koptu. Normal internet geri geldi.',

    download: 'İndirme', upload: 'Yükleme', device: 'Telefon', publicIp: 'Genel IP', uptime: 'Süre',
    totalData: 'Kullanılan veri', connections: 'Bağlantılar', none: '—', anyPhone: 'Herhangi bir telefon',
    phonesFound: 'Telefonlar', noPhones: 'Bağlı telefon yok',
    stUnauthorized: 'izin bekliyor', stOffline: 'çevrim dışı', stDevice: 'hazır',

    hsTitle: 'Mobil etkin nokta', hsDesc: 'Tüneli Wi-Fi ile başka cihazlarla paylaş. Windows Mobil etkin nokta kullanılır.',
    hsOn: 'Etkin nokta açık', hsOff: 'Etkin nokta kapalı', hsBusy: 'Çalışıyor…',
    hsStart: 'Aç', hsStop: 'Kapat', hsName: 'Ağ adı', hsPass: 'Şifre',
    hsClients: 'Bağlı cihaz', hsCustom: 'Özel ad ve şifre', hsCustomDesc: 'Boş bırakırsan Windows ayarı kalır.',
    hsWithTun: 'Tünelle birlikte aç', hsWithTunDesc: 'Tünel her bağlandığında etkin noktayı aç.',
    hsTip: 'Önce tüneli bağla. Sonra etkin nokta telefonun internetini paylaşır.',
    copy: 'Kopyala', copied: 'Kopyalandı', show: 'Göster', hide: 'Gizle', save: 'Kaydet', saved: 'Kaydedildi',

    gTitle: 'TelTun nasıl çalışır', gLead: 'TelTun, telefonun mobil internetini USB kablo ile bilgisayara verir. Adım adım ne olduğu burada.',
    gSetupTitle: 'İlk kurulum (sadece bir kez)',
    g1t: 'Geliştirici seçeneklerini aç', g1d: 'Ayarlar → Telefon hakkında → "Yapı numarası"na (veya "MIUI / OS sürümü") 7 kez dokun.',
    g2t: 'USB hata ayıklamayı aç', g2d: 'Ayarlar → Ek ayarlar → Geliştirici seçenekleri → USB hata ayıklama.',
    g3t: 'Kabloyu tak', g3d: 'İyi bir veri kablosu kullan. Telefon sorarsa "Dosya aktarımı" seç.',
    g4t: 'Bu bilgisayara izin ver', g4d: 'Telefonun kilidini aç, USB hata ayıklama sorusunda "İzin ver"e dokun. "Her zaman izin ver"i işaretle.',
    gFlowTitle: 'Verinin yolu',
    fApps: 'PC uygulamaları', fAppsD: 'Tarayıcı, oyun, güncelleme',
    fTun: 'TelTun kartı', fTunD: 'Sanal ağ kartı tüm trafiği alır',
    fUsb: 'USB hattı', fUsbD: 'Tek adb akışı, içinde çok bağlantı',
    fPhone: 'Telefon yardımcısı', fPhoneD: 'Gerçek bağlantıları açar',
    fNet: 'Mobil ağ', fNetD: '4G / 5G internet',
    gWhyTitle: 'Neden tek akış?',
    gWhy: 'adb tek akışta yaklaşık 300 Mbps taşır. Ama aynı anda çok akış açılınca birbirini yavaşlatır. Bu yüzden TelTun bütün bağlantıları tek akışın içine koyar (çoğullama). Hızın sırrı bu.',
    gAdminTitle: 'Neden yönetici izni?',
    gAdmin: 'Sanal ağ kartı kurmak ve yönlendirmeyi değiştirmek yönetici izni ister. TelTun açılırken bir kez sorar.',
    gSafeTitle: 'Güvenli mi?',
    gSafe: 'Telefondaki yardımcı sadece telefonun kendisini dinler (127.0.0.1). Ağa açık bir şey yok. Kesince ya da kapatınca sanal kart silinir, normal internet geri gelir.',
    gTroubleTitle: 'Sorun çıkarsa',
    gTrouble: 'Bağlantıyı kes, kabloyu çıkar, geri tak ve yeniden bağlan. Ayrıntı için Kayıtlar sayfasına bak.',

    logsTitle: 'Kayıtlar', logsOpen: 'Klasörü aç', logsCopy: 'Hepsini kopyala', logsEmpty: 'Henüz bir şey yok.',

    setTitle: 'Ayarlar', setLang: 'Dil', setLangDesc: 'Arayüz dili.',
    setAutostart: 'Windows ile başlat', setAutostartDesc: 'Oturum açınca tepside açılır.',
    setAutoConnect: 'TelTun açılınca bağlan', setAutoConnectDesc: 'Telefon bulunur bulunmaz bağlan.',
    setAutoReconnect: 'Kopunca yeniden bağlan', setAutoReconnectDesc: 'Kablo gevşerse telefon geri gelince yeniden bağlan.',
    setTray: 'Tepsiye küçült', setTrayDesc: 'X düğmesi pencereyi gizler. Çıkış tepsi menüsünden.',
    setQuic: 'QUIC engelle (UDP 443)', setQuicDesc: 'Çoğu sitede daha hızlı. Tarayıcı normal HTTPS\'e geçer.',
    setDevice: 'Kullanılacak telefon', setDeviceDesc: 'Birden çok telefon takarsan birini seç.',
    setApplyNext: 'Sonraki bağlantıda geçerli olur.',

    abTitle: 'Hakkında', abDesc: 'Telefonun mobil verisini USB ile Windows bilgisayarla paylaş. Ücretsiz ve açık kaynak, öğrenmek için yapıldı.',
    abMadeBy: 'Made by Enes', abVersion: 'Sürüm', abLicense: 'Lisans', abSource: 'Kaynak kod',
    abCredits: 'Kullanılanlar', abNote: 'Kendi telefonun ve hattınla kullan, operatörünün kurallarına uy.',
  },
};

let lang = 'en';
export function setLang(l) { lang = dict[l] ? l : 'en'; document.documentElement.lang = lang; }
export function getLang() { return lang; }
export function t(key) { return (dict[lang] && dict[lang][key]) ?? dict.en[key] ?? key; }
export function apply(root = document) {
  root.querySelectorAll('[data-i18n]').forEach(el => { el.textContent = t(el.dataset.i18n); });
  root.querySelectorAll('[data-i18n-title]').forEach(el => { el.title = t(el.dataset.i18nTitle); });
  root.querySelectorAll('[data-i18n-ph]').forEach(el => { el.placeholder = t(el.dataset.i18nPh); });
}
