"use client";

import { motion } from "framer-motion";

export default function CallToAction() {
  return (
    <section className="py-24 relative overflow-hidden z-10">
      <div className="absolute inset-0 bg-gradient-to-b from-transparent to-blue-900/10" />
      <div className="container mx-auto px-4 relative z-10">
        <motion.div 
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="glass-panel max-w-5xl mx-auto p-12 lg:p-20 text-center relative overflow-hidden"
        >
          {/* Subtle glow inside the CTA */}
          <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[300px] h-[300px] bg-blue-500/20 rounded-full blur-[80px]" />
          
          <h2 className="text-4xl lg:text-5xl font-bold mb-6 relative z-10">
            Ready to streamline your workflow?
          </h2>
          <p className="text-xl text-gray-400 mb-10 max-w-2xl mx-auto relative z-10">
            Join thousands of modern companies using Orbit to manage their employees and appointments seamlessly.
          </p>
          <button className="relative z-10 px-10 py-5 bg-white text-black font-bold rounded-full hover:bg-gray-200 transition-colors shadow-[0_0_30px_rgba(255,255,255,0.2)]">
            Start Your Free Trial
          </button>
        </motion.div>
      </div>
    </section>
  );
}
