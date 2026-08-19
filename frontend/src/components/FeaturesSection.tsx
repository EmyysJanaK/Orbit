"use client";

import { motion } from "framer-motion";
import { Calendar, CreditCard, BarChart3, Users } from "lucide-react";

const features = [
  {
    icon: <Calendar className="w-6 h-6 text-blue-400" />,
    title: "Smart Appointments",
    description: "Schedule, manage, and track employee appointments with ease and automation.",
  },
  {
    icon: <CreditCard className="w-6 h-6 text-purple-400" />,
    title: "Integrated Payments",
    description: "Seamless Stripe billing, automated payment workflows, and invoicing.",
  },
  {
    icon: <BarChart3 className="w-6 h-6 text-green-400" />,
    title: "Real-time Analytics",
    description: "Make data-driven decisions with real-time usage insights and reporting.",
  },
  {
    icon: <Users className="w-6 h-6 text-pink-400" />,
    title: "Team Management",
    description: "Secure, role-based access control and comprehensive user management.",
  }
];

export default function FeaturesSection() {
  return (
    <section className="py-24 relative z-10">
      <div className="container mx-auto px-4">
        <div className="text-center mb-20">
          <motion.h2 
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="text-3xl lg:text-5xl font-bold mb-6"
          >
            Everything you need
          </motion.h2>
          <motion.p 
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.1 }}
            className="text-gray-400 text-lg max-w-2xl mx-auto"
          >
            Powerful tools designed to help your business scale efficiently without the administrative overhead.
          </motion.p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 max-w-7xl mx-auto">
          {features.map((feature, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-100px" }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              whileHover={{ y: -5, scale: 1.02 }}
              className="glass-panel p-8 flex flex-col items-start transition-all cursor-pointer group"
            >
              <div className="p-4 bg-white/5 rounded-xl mb-6 group-hover:bg-white/10 transition-colors">
                {feature.icon}
              </div>
              <h3 className="text-xl font-semibold mb-3">{feature.title}</h3>
              <p className="text-gray-400 leading-relaxed text-sm">
                {feature.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
