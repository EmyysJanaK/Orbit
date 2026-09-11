import Link from "next/link";
import HeroSection from "@/components/HeroSection";
import FeaturesSection from "@/components/FeaturesSection";
import CallToAction from "@/components/CallToAction";

export default function Home() {
  return (
    <main className="min-h-screen selection:bg-blue-500/30">
      {/* Navigation Bar */}
      <nav className="w-full fixed top-0 z-50 glass-panel border-x-0 border-t-0 rounded-none bg-black/40 backdrop-blur-md">
        <div className="container mx-auto px-6 h-20 flex items-center justify-between">
          <div className="text-2xl font-black tracking-tighter flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500" />
            ORBIT
          </div>
          <div className="flex gap-4 items-center">
            <Link href="/login" className="px-5 py-2 text-sm font-medium text-gray-300 hover:text-white transition-colors">
              Login
            </Link>
            <Link href="/signup" className="px-5 py-2 text-sm font-medium bg-white text-black rounded-full hover:bg-gray-200 transition-colors">
              Sign Up
            </Link>
          </div>
        </div>
      </nav>

      <HeroSection />
      <FeaturesSection />
      <CallToAction />
      
      <footer className="py-12 border-t border-white/10 mt-10 relative z-10">
        <div className="container mx-auto px-6 text-center text-gray-500 text-sm">
          © {new Date().getFullYear()} Orbit Employee Management. All rights reserved.
        </div>
      </footer>
    </main>
  );
}
