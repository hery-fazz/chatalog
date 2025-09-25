'use client';

import { useState, useEffect } from 'react';

// src/app/page.tsx — Slideshow percakapan WA dengan nama toko/brand, slide ketiga jadi usaha cuci motor

const WA = process.env.NEXT_PUBLIC_WA_NUMBER || "62xxxxxxxxxxx"; // set di .env.local
const msg = encodeURIComponent("Halo! Saya ingin coba Chatalog.");
const waUrl = `https://wa.me/${WA}?text=${msg}`;

const chats = [
  {
    user: "Halo, saya Toko Ponsel Jaya. Buatkan brosur promo HP cicilan mulai 300rb/bln",
    bot: ["Siap! Berikut brosurnya:", "• Smartphone 4G", "• Cicilan Rp300.000/bln", "• Bonus kuota 10GB"],
    note: "(Foto produk HP)",
  },
  {
    user: "Hai, Laundry Bersih ingin promo paket 5kg 25rb",
    bot: ["Tentu! Ini kartunya:", "• Laundry Express 5kg", "• Rp25.000", "• Selesai 4 jam"],
    note: "(Foto laundry)",
  },
  {
    user: "Halo, Cuci Motor Kilat mau promo paket cuci motor 15rb",
    bot: ["Oke, ini hasilnya:", "• Cuci Motor Kilat", "• Rp15.000", "• Gratis semir ban", "• Estimasi 15 menit"],
    note: "(Foto cuci motor)",
  },
];

// Dummy image previews: coba pakai file nyata di /public lebih dulu, fallback ke SVG dummy
const images = [
  {
    alt: "Smartphone 4G",
    path: "/promo_card_1.png", // letakkan di public/promo_card_1.png
    fallback:
      "data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='640' height='360'><defs><linearGradient id='g' x1='0' x2='1'><stop stop-color='%2300a89d' offset='0'/><stop stop-color='%2300665f' offset='1'/></linearGradient></defs><rect width='100%' height='100%' fill='url(%23g)'/><rect x='260' y='40' rx='16' ry='16' width='120' height='280' fill='rgba(255,255,255,0.2)'/><text x='50%25' y='50%25' fill='white' font-size='24' text-anchor='middle'>Smartphone</text></svg>",
  },
  {
    alt: "Laundry 5kg",
    path: "/promo_card_2.png", // letakkan di public/promo_card_2.png
    fallback:
      "data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='640' height='360'><defs><linearGradient id='g2' x1='0' x2='1'><stop stop-color='%2300e0d2' offset='0'/><stop stop-color='%23007e76' offset='1'/></linearGradient></defs><rect width='100%' height='100%' fill='url(%23g2)'/><circle cx='320' cy='180' r='80' fill='rgba(255,255,255,0.2)'/><text x='50%25' y='50%25' fill='white' font-size='24' text-anchor='middle'>Laundry</text></svg>",
  },
  {
    alt: "Cuci Motor",
    path: "/promo_card_3.png", // letakkan di public/promo_card_3.png
    fallback:
      "data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='640' height='360'><defs><linearGradient id='g3' x1='0' x2='1'><stop stop-color='%23ff6f61' offset='0'/><stop stop-color='%23a83a30' offset='1'/></linearGradient></defs><rect width='100%' height='100%' fill='url(%23g3)'/><rect x='180' y='190' width='280' height='40' rx='20' fill='rgba(255,255,255,0.2)'/><circle cx='240' cy='250' r='20' fill='rgba(255,255,255,0.35)'/><circle cx='380' cy='250' r='20' fill='rgba(255,255,255,0.35)'/><text x='50%25' y='40%25' fill='white' font-size='24' text-anchor='middle'>Cuci Motor</text></svg>",
  },
];

export default function Page() {
  const [current, setCurrent] = useState(0);
  const [paused, setPaused] = useState(false);

  // autoplay: ganti slide tiap 4 detik, pause saat hover/touch
  useEffect(() => {
    if (paused) return;
    const id = setInterval(() => {
      setCurrent((c) => (c + 1) % chats.length);
    }, 4000);
    return () => clearInterval(id);
  }, [paused]);

  return (
    <main className="min-h-screen bg-[#0B0D10] text-white">
      {/* Sticky mobile CTA */}
      <div className="fixed inset-x-0 bottom-0 z-40 border-t border-white/10 bg-[#0B0D10]/95 backdrop-blur md:hidden">
        <div className="mx-auto max-w-5xl px-4 py-3">
          <a
            href={waUrl}
            className="block w-full rounded-xl bg-[#00A89D] px-4 py-3 text-center text-sm font-semibold hover:opacity-95 active:opacity-90"
          >
            Chat WhatsApp
          </a>
        </div>
      </div>

      {/* Header */}
      <header className="sticky top-0 z-50 border-b border-white/10 bg-[#0B0D10]/90 backdrop-blur">
        <div className="mx-auto max-w-5xl px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-2 sm:gap-3">
            <div className="size-7 sm:size-8 rounded-xl bg-[#00A89D] grid place-items-center relative">
              <span className="text-[10px] sm:text-xs font-black tracking-wider">CT</span>
              <span className="absolute -right-1 -bottom-1 size-2 rounded-full bg-[#FF6F61]" />
            </div>
            <span className="text-sm sm:text-base font-semibold">Chatalog</span>
          </div>
          <a
            href={waUrl}
            className="hidden md:inline-flex rounded-xl px-3.5 py-2 text-sm font-semibold bg-[#FF6F61] hover:opacity-95"
          >
            Chat WhatsApp
          </a>
        </div>
      </header>

      {/* Hero */}
      <section className="relative overflow-hidden">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(400px_circle_at_10%_-10%,rgba(0,168,157,0.25),transparent),radial-gradient(400px_circle_at_90%_-20%,rgba(255,111,97,0.2),transparent)] md:bg-[radial-gradient(600px_circle_at_10%_-10%,rgba(0,168,157,0.25),transparent),radial-gradient(600px_circle_at_90%_-20%,rgba(255,111,97,0.2),transparent)]" />
        <div className="relative mx-auto max-w-5xl px-4 pt-8 pb-24 md:pt-14 md:pb-24">
          <div className="grid md:grid-cols-2 gap-6 md:gap-10 items-center">
            <div>
              <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-2.5 py-1 text-[11px] text-white/80 mb-4 md:mb-5">
                <span className="size-1.5 rounded-full bg-[#00A89D]" />
                Media promosi instan dari chat
              </div>
              <h1 className="text-3xl sm:text-4xl md:text-5xl font-bold leading-[1.15]">
                Katalog & Brosur Online dari <span className="text-[#00A89D]">Chat WhatsApp</span>
              </h1>
              <p className="mt-3 md:mt-4 text-white/70 text-base sm:text-lg">
                Dari laundry, cuci motor, kios pulsa, hingga warung makan—Chatalog bantu buat brosur online & katalog produk otomatis.
              </p>
              <div className="mt-6 md:mt-8 grid grid-cols-1 sm:grid-cols-2 gap-3">
                <a href={waUrl} className="rounded-xl px-4 py-3 text-center font-semibold bg-[#00A89D] hover:opacity-95">
                  Mulai via WhatsApp
                </a>
                <a href="#how" className="rounded-xl px-4 py-3 text-center font-semibold bg-white text-[#0B0D10] hover:opacity-90">
                  Cara pakai
                </a>
              </div>
              <p className="mt-3 text-xs text-white/50">Tidak perlu install aplikasi tambahan.</p>
            </div>

            {/* Slideshow percakapan WA */}
            <div className="relative">
              <div
                className="mx-auto w-full max-w-[260px] sm:max-w-[280px] rounded-xl border border-white/10 bg-white/5 p-3 overflow-hidden"
                onMouseEnter={() => setPaused(true)}
                onMouseLeave={() => setPaused(false)}
                onTouchStart={() => setPaused(true)}
                onTouchEnd={() => setPaused(false)}
              >
                <div className="flex items-center justify-between rounded-lg border border-white/10 bg-[#0B0D10] px-2.5 py-2">
                  <div className="flex items-center gap-2">
                    <div className="size-5 rounded-full bg-[#25D366]" />
                    <div className="text-[11px] text-white/80">WA • Chatalog</div>
                  </div>
                  <div className="text-[9px] text-white/50">Online</div>
                </div>

                {/* Slide content */}
                <div className="mt-3 space-y-2 text-[12px]">
                  {/* User bubble */}
                  <div className="flex justify-end">
                    <div className="max-w-[75%] rounded-2xl rounded-br-sm bg-[#005C4B] px-3 py-2">
                      {chats[current].user}
                    </div>
                  </div>
                  {/* Bot bubble */}
                  <div className="flex justify-start">
                    <div className="max-w-[80%] rounded-2xl rounded-bl-sm bg-white/10 px-3 py-2">
                      <div className="opacity-90">{chats[current].bot[0]}</div>
                      <div className="mt-1 text-white/80">
                        {chats[current].bot.slice(1).map((line, idx) => (
                          <div key={idx}>{line}</div>
                        ))}
                      </div>
                      <div className="mt-2 overflow-hidden rounded-md border border-white/10">
                        <img
                          src={images[current].path}
                          alt={images[current].alt}
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).src = images[current].fallback;
                          }}
                          className="w-full h-auto max-h-28 object-contain bg-black/40"
                        />
                      </div>
                      
                    </div>
                  </div>
                </div>

                {/* Indicator dots */}
                <div className="mt-4 flex justify-center gap-2">
                  {chats.map((_, i) => (
                    <button
                      key={i}
                      onClick={() => setCurrent(i)}
                      className={`size-2.5 rounded-full ${i === current ? 'bg-[#00A89D]' : 'bg-white/30'}`}
                      aria-label={`Slide ${i+1}`}
                      aria-current={i === current}
                    />
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Cara pakai */}
      <section id="how" className="border-t border-white/10">
        <div className="mx-auto max-w-5xl px-4 py-10 md:py-12">
          <h2 className="text-xl sm:text-2xl md:text-3xl font-bold">Cara pakai</h2>
          <ol className="mt-4 grid gap-3 sm:grid-cols-3 text-sm text-white/80">
            <li className="rounded-xl border border-white/10 bg-white/5 p-4"><b>1) Chat</b><br/>Kirim detail produk/jasa di WhatsApp.</li>
            <li className="rounded-xl border border-white/10 bg-white/5 p-4"><b>2) Rapi otomatis</b><br/>Jadi katalog atau brosur online siap share.</li>
            <li className="rounded-xl border border-white/10 bg-white/5 p-4"><b>3) Bagikan</b><br/>Sebar via WA/IG/FB atau cetak QR di toko.</li>
          </ol>
        </div>
      </section>

      {/* Footer */}
      <footer className="pb-16 md:pb-0 border-t border-white/10">
        <div className="mx-auto max-w-5xl px-4 py-8 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs sm:text-[13px] text-white/60">
          <div className="flex items-center gap-2 sm:gap-3">
            <div className="size-7 sm:size-8 rounded-xl bg-[#00A89D] grid place-items-center relative">
              <span className="text-[10px] sm:text-xs font-black tracking-wider">CT</span>
              <span className="absolute -right-1 -bottom-1 size-2 rounded-full bg-[#FF6F61]" />
            </div>
            <div>
              <div className="font-semibold text-white/80">Chatalog</div>
              <div className="text-[10px] sm:text-xs">Media promosi instan dari chat</div>
            </div>
          </div>
          <div>© {new Date().getFullYear()} Chatalog</div>
        </div>
      </footer>
    </main>
  );
}
