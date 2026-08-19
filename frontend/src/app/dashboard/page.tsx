"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { 
  Calendar, 
  Plus, 
  Trash2, 
  LogOut, 
  CheckCircle, 
  Clock, 
  Loader2,
  User as UserIcon 
} from "lucide-react";
import api from "@/lib/api";
import CreateAppointmentModal from "@/components/CreateAppointmentModal";
import Link from "next/link";

interface Appointment {
  id: number;
  user_id: number;
  title: string;
  scheduled_at: string;
  status: string;
  created_at: string;
}

export default function DashboardPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [user, setUser] = useState<{ email: string; role: string } | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    const storedToken = localStorage.getItem("token");
    const storedUser = localStorage.getItem("user");
    if (!storedToken || !storedUser) {
      router.push("/login");
    } else {
      try {
        setUser(JSON.parse(storedUser));
      } catch (e) {
        router.push("/login");
      }
    }
  }, [router]);

  const { data: appointments = [], isLoading, isError } = useQuery<Appointment[]>({
    queryKey: ["appointments"],
    queryFn: async () => {
      const { data } = await api.get("/api/appointments");
      return data;
    },
    enabled: !!user,
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/api/appointments/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["appointments"] });
    },
  });

  const statusMutation = useMutation({
    mutationFn: async ({ id, status }: { id: number; status: string }) => {
      await api.patch(`/api/appointments/${id}`, { status });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["appointments"] });
    },
  });

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    router.push("/login");
  };

  if (!user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  const scheduledCount = appointments.filter((a) => a.status === "scheduled").length;
  const completedCount = appointments.filter((a) => a.status === "completed").length;

  return (
    <div className="min-h-screen pb-12 selection:bg-blue-500/30">
      {/* Top Navbar */}
      <nav className="w-full glass-panel border-x-0 border-t-0 rounded-none bg-black/40 backdrop-blur-md">
        <div className="container mx-auto px-6 h-20 flex items-center justify-between">
          <Link href="/" className="text-2xl font-black tracking-tighter flex items-center gap-2">
            <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500" />
            ORBIT
          </Link>

          <div className="flex items-center gap-6">
            <div className="flex items-center gap-3 px-4 py-2 glass-panel rounded-full">
              <UserIcon size={16} className="text-blue-400" />
              <span className="text-sm font-medium">{user.email}</span>
              <span className="text-xs px-2 py-0.5 rounded bg-blue-500/20 text-blue-300 font-semibold uppercase">
                {user.role}
              </span>
            </div>
            
            <button
              onClick={handleLogout}
              className="p-2 text-gray-400 hover:text-white transition-colors"
              title="Logout"
            >
              <LogOut size={20} />
            </button>
          </div>
        </div>
      </nav>

      {/* Main Content Area */}
      <main className="container mx-auto px-6 pt-10">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-8 gap-4">
          <div>
            <h1 className="text-3xl font-extrabold mb-1">Appointments Dashboard</h1>
            <p className="text-gray-400 text-sm">Manage and track your scheduled meetings and tasks</p>
          </div>

          <button
            onClick={() => setIsModalOpen(true)}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-lg transition-colors flex items-center gap-2 shadow-[0_0_20px_rgba(37,99,235,0.3)]"
          >
            <Plus size={20} /> New Appointment
          </button>
        </div>

        {/* Metrics Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-10">
          <div className="glass-panel p-6 flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400 font-medium">Total Appointments</p>
              <h3 className="text-3xl font-bold mt-1">{appointments.length}</h3>
            </div>
            <div className="p-3 bg-blue-500/10 text-blue-400 rounded-xl">
              <Calendar size={24} />
            </div>
          </div>

          <div className="glass-panel p-6 flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400 font-medium">Scheduled / Upcoming</p>
              <h3 className="text-3xl font-bold mt-1">{scheduledCount}</h3>
            </div>
            <div className="p-3 bg-amber-500/10 text-amber-400 rounded-xl">
              <Clock size={24} />
            </div>
          </div>

          <div className="glass-panel p-6 flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-400 font-medium">Completed</p>
              <h3 className="text-3xl font-bold mt-1">{completedCount}</h3>
            </div>
            <div className="p-3 bg-emerald-500/10 text-emerald-400 rounded-xl">
              <CheckCircle size={24} />
            </div>
          </div>
        </div>

        {/* Table View */}
        <div className="glass-panel overflow-hidden">
          {isLoading ? (
            <div className="p-12 text-center text-gray-400 flex items-center justify-center gap-3">
              <Loader2 className="animate-spin text-blue-500" size={24} /> Loading appointments...
            </div>
          ) : isError ? (
            <div className="p-12 text-center text-red-400">
              Failed to load appointments. Please check if your Go backend is running.
            </div>
          ) : appointments.length === 0 ? (
            <div className="p-16 text-center">
              <Calendar size={48} className="mx-auto text-gray-600 mb-4" />
              <h3 className="text-lg font-semibold mb-1">No appointments found</h3>
              <p className="text-gray-400 text-sm mb-6">Create your first appointment to get started.</p>
              <button
                onClick={() => setIsModalOpen(true)}
                className="px-5 py-2.5 bg-white text-black font-semibold rounded-lg hover:bg-gray-200 transition-colors"
              >
                Create Appointment
              </button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className="border-b border-white/10 text-gray-400 text-xs uppercase tracking-wider bg-white/5">
                    <th className="p-4 pl-6">Title</th>
                    <th className="p-4">Scheduled Date</th>
                    <th className="p-4">Status</th>
                    <th className="p-4 pr-6 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5 text-sm">
                  {appointments.map((item) => (
                    <tr key={item.id} className="hover:bg-white/5 transition-colors">
                      <td className="p-4 pl-6 font-medium text-white">{item.title}</td>
                      <td className="p-4 text-gray-300">
                        {new Date(item.scheduled_at).toLocaleString(undefined, {
                          dateStyle: "medium",
                          timeStyle: "short",
                        })}
                      </td>
                      <td className="p-4">
                        <select
                          value={item.status}
                          onChange={(e) =>
                            statusMutation.mutate({ id: item.id, status: e.target.value })
                          }
                          className="bg-black/60 border border-white/10 text-xs font-semibold rounded-full px-3 py-1 text-white focus:outline-none cursor-pointer"
                        >
                          <option value="scheduled">Scheduled</option>
                          <option value="pending">Pending</option>
                          <option value="completed">Completed</option>
                          <option value="cancelled">Cancelled</option>
                        </select>
                      </td>
                      <td className="p-4 pr-6 text-right">
                        <button
                          onClick={() => deleteMutation.mutate(item.id)}
                          className="p-2 text-gray-400 hover:text-red-400 transition-colors"
                          title="Delete appointment"
                        >
                          <Trash2 size={18} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </main>

      <CreateAppointmentModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </div>
  );
}
