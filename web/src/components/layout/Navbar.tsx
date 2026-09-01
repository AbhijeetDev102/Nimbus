import React from "react";
import { CloudLightning, LayoutDashboard, Sparkles, Terminal, RefreshCw } from "lucide-react";

interface NavbarProps {
  activeTab: "overview" | "studio" | "executions";
  setActiveTab: (tab: "overview" | "studio" | "executions") => void;
  onRefresh?: () => void;
  isRefreshing?: boolean;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  onRefresh,
  isRefreshing,
}) => {
  return (
    <header className="sticky top-0 z-40 w-full border-b border-white/10 bg-[#06090e]/80 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500 via-indigo-500 to-purple-500 p-[1px] shadow-lg shadow-cyan-500/20">
            <div className="w-full h-full bg-[#080d16] rounded-xl flex items-center justify-center">
              <CloudLightning className="w-5 h-5 text-cyan-400" />
            </div>
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="text-xl font-bold tracking-tight text-white">Nimbus</span>
              <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-cyan-950 text-cyan-400 border border-cyan-800/60">
                v1.0
              </span>
            </div>
            <p className="text-xs text-slate-400 hidden sm:block">Distributed Job Execution Engine</p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="flex items-center space-x-1 sm:space-x-2 bg-slate-900/60 p-1 rounded-xl border border-white/5">
          <button
            onClick={() => setActiveTab("overview")}
            className={`flex items-center space-x-2 px-3 sm:px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
              activeTab === "overview"
                ? "bg-cyan-500/10 text-cyan-400 border border-cyan-500/30 shadow-sm"
                : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
            }`}
          >
            <LayoutDashboard className="w-4 h-4" />
            <span>Overview</span>
          </button>

          <button
            onClick={() => setActiveTab("studio")}
            className={`flex items-center space-x-2 px-3 sm:px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
              activeTab === "studio"
                ? "bg-purple-500/10 text-purple-400 border border-purple-500/30 shadow-sm"
                : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
            }`}
          >
            <Sparkles className="w-4 h-4" />
            <span>Studio</span>
          </button>

          <button
            onClick={() => setActiveTab("executions")}
            className={`flex items-center space-x-2 px-3 sm:px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
              activeTab === "executions"
                ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 shadow-sm"
                : "text-slate-400 hover:text-slate-200 hover:bg-white/5"
            }`}
          >
            <Terminal className="w-4 h-4" />
            <span>Executions</span>
          </button>
        </nav>

        {/* Right Status Indicator & Refresh */}
        <div className="flex items-center space-x-3">
          <div className="hidden md:flex items-center space-x-2 px-3 py-1 rounded-full bg-slate-900/80 border border-emerald-500/20 text-xs text-emerald-400 font-mono">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span>Gateway :8081</span>
          </div>

          {onRefresh && (
            <button
              onClick={onRefresh}
              disabled={isRefreshing}
              className="p-2 rounded-lg bg-slate-900/60 border border-white/10 text-slate-400 hover:text-white hover:bg-slate-800 transition-colors disabled:opacity-50 cursor-pointer"
              title="Refresh Stats"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? "animate-spin text-cyan-400" : ""}`} />
            </button>
          )}
        </div>
      </div>
    </header>
  );
};
