"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  ShieldCheck,
  Building2,
  Layers,
  Grid3X3,
  Plus,
  Trash2,
  Edit2,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  Sparkles,
  MapPin,
  Users,
  Settings,
  Palette,
  Eye,
  Lock,
  ArrowRight,
  ChevronRight,
  Info,
} from "lucide-react";
import {
  fetchVenues,
  createVenue,
  updateVenue,
  deleteVenue,
  fetchCategories,
  createCategory,
  fetchVenueSeats,
  batchConfigureSeats,
  deleteVenueSeats,
} from "../../lib/api";
import {
  Venue,
  SeatCategory,
  VenueSeat,
  CreateSeatPayload,
} from "../../lib/types";
import { useAuth } from "../../context/AuthContext";

export default function AdminPage() {
  const { user, isLoading: authLoading, switchRole } = useAuth();

  // Navigation tab: "venues" | "categories" | "layout"
  const [activeTab, setActiveTab] = useState<"venues" | "categories" | "layout">("venues");

  // Data states
  const [venues, setVenues] = useState<Venue[]>([]);
  const [categories, setCategories] = useState<SeatCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Venue Modal state
  const [showVenueModal, setShowVenueModal] = useState(false);
  const [editingVenue, setEditingVenue] = useState<Venue | null>(null);
  const [venueName, setVenueName] = useState("");
  const [venueAddress, setVenueAddress] = useState("");
  const [venueCity, setVenueCity] = useState("");
  const [venueState, setVenueState] = useState("");
  const [venueCountry, setVenueCountry] = useState("IN");
  const [savingVenue, setSavingVenue] = useState(false);

  // Category Modal state
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [categoryName, setCategoryName] = useState("");
  const [categoryDesc, setCategoryDesc] = useState("");
  const [categoryColor, setCategoryColor] = useState("#3B82F6");
  const [savingCategory, setSavingCategory] = useState(false);

  // Seat Layout Builder state
  const [selectedVenueId, setSelectedVenueId] = useState<string>("");
  const [venueSeats, setVenueSeats] = useState<VenueSeat[]>([]);
  const [loadingSeats, setLoadingSeats] = useState(false);

  // Layout Generator Controls
  const [gridRowsCount, setGridRowsCount] = useState<number>(6);
  const [gridColsCount, setGridColsCount] = useState<number>(10);
  const [rowCategoryMapping, setRowCategoryMapping] = useState<Record<string, string>>({});
  const [customSeats, setCustomSeats] = useState<CreateSeatPayload[]>([]);
  const [deployingLayout, setDeployingLayout] = useState(false);

  const rowLetters = ["A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"];

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [vList, cList] = await Promise.all([fetchVenues(), fetchCategories()]);
      setVenues(vList);
      setCategories(cList);

      if (vList.length > 0 && !selectedVenueId) {
        setSelectedVenueId(vList[0].id);
        loadVenueSeats(vList[0].id);
      }
    } catch (err: any) {
      setError(err.message || "Failed to load admin resources");
    } finally {
      setLoading(false);
    }
  };

  const loadVenueSeats = async (venueId: string) => {
    if (!venueId) return;
    setLoadingSeats(true);
    try {
      const seats = await fetchVenueSeats(venueId);
      setVenueSeats(seats);
    } catch {
      setVenueSeats([]);
    } finally {
      setLoadingSeats(false);
    }
  };

  useEffect(() => {
    if (!authLoading) {
      loadData();
    }
  }, [authLoading]);

  useEffect(() => {
    if (selectedVenueId) {
      loadVenueSeats(selectedVenueId);
    }
  }, [selectedVenueId]);

  // Pre-fill initial row category map when categories load
  useEffect(() => {
    if (categories.length > 0) {
      const mapping: Record<string, string> = {};
      const vipCat = categories.find((c) => c.name.toLowerCase().includes("vip")) || categories[0];
      const premCat = categories.find((c) => c.name.toLowerCase().includes("premium")) || categories[0];
      const stdCat = categories.find((c) => c.name.toLowerCase().includes("standard")) || categories[categories.length - 1];

      rowLetters.slice(0, gridRowsCount).forEach((row, idx) => {
        if (idx === 0 && vipCat) {
          mapping[row] = vipCat.id;
        } else if (idx <= 2 && premCat) {
          mapping[row] = premCat.id;
        } else {
          mapping[row] = stdCat.id;
        }
      });
      setRowCategoryMapping(mapping);
    }
  }, [categories, gridRowsCount]);

  // Generate matrix seats whenever grid dimensions or row mappings change
  useEffect(() => {
    if (categories.length === 0) return;
    const defaultCatId = categories[0]?.id || "";
    const generated: CreateSeatPayload[] = [];

    const activeRows = rowLetters.slice(0, Math.min(gridRowsCount, rowLetters.length));
    activeRows.forEach((rowLabel, rIdx) => {
      const catId = rowCategoryMapping[rowLabel] || defaultCatId;
      for (let c = 1; c <= gridColsCount; c++) {
        generated.push({
          seat_category_id: catId,
          row_label: rowLabel,
          seat_number: c.toString(),
          grid_row: rIdx + 1,
          grid_col: c,
          is_active: true,
        });
      }
    });
    setCustomSeats(generated);
  }, [gridRowsCount, gridColsCount, rowCategoryMapping, categories]);

  // Handlers for Venue CRUD
  const handleOpenVenueModal = (venue?: Venue) => {
    if (venue) {
      setEditingVenue(venue);
      setVenueName(venue.name);
      setVenueAddress(venue.address);
      setVenueCity(venue.city);
      setVenueState(venue.state || "");
      setVenueCountry(venue.country);
    } else {
      setEditingVenue(null);
      setVenueName("");
      setVenueAddress("");
      setVenueCity("");
      setVenueState("");
      setVenueCountry("IN");
    }
    setShowVenueModal(true);
  };

  const handleSaveVenue = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!venueName.trim() || !venueAddress.trim() || !venueCity.trim()) {
      alert("Name, Address, and City are required.");
      return;
    }

    setSavingVenue(true);
    setError(null);
    try {
      if (editingVenue) {
        await updateVenue(editingVenue.id, {
          name: venueName.trim(),
          address: venueAddress.trim(),
          city: venueCity.trim(),
          state: venueState.trim() || undefined,
          country: venueCountry.trim() || "IN",
          total_capacity: editingVenue.total_capacity,
        });
        setSuccessMessage(`Venue "${venueName}" updated successfully.`);
      } else {
        const created = await createVenue({
          name: venueName.trim(),
          address: venueAddress.trim(),
          city: venueCity.trim(),
          state: venueState.trim() || undefined,
          country: venueCountry.trim() || "IN",
        });
        setSuccessMessage(`Venue "${created.name}" created successfully.`);
      }
      setShowVenueModal(false);
      await loadData();
    } catch (err: any) {
      alert(`Error saving venue: ${err.message}`);
    } finally {
      setSavingVenue(false);
    }
  };

  const handleDeleteVenue = async (venue: Venue) => {
    if (!window.confirm(`Are you sure you want to delete venue "${venue.name}"? All associated seats will be removed.`)) {
      return;
    }
    try {
      await deleteVenue(venue.id);
      setSuccessMessage(`Venue "${venue.name}" deleted.`);
      await loadData();
    } catch (err: any) {
      alert(`Failed to delete venue: ${err.message}`);
    }
  };

  // Handlers for Category CRUD
  const handleSaveCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!categoryName.trim()) {
      alert("Category name is required.");
      return;
    }

    setSavingCategory(true);
    try {
      const created = await createCategory({
        name: categoryName.trim(),
        description: categoryDesc.trim() || undefined,
        color_code: categoryColor,
      });
      setSuccessMessage(`Seat category "${created.name}" created successfully.`);
      setShowCategoryModal(false);
      setCategoryName("");
      setCategoryDesc("");
      await loadData();
    } catch (err: any) {
      alert(`Error creating category: ${err.message}`);
    } finally {
      setSavingCategory(false);
    }
  };

  // Handlers for Layout Configurator
  const handleToggleSeatActive = (index: number) => {
    setCustomSeats((prev) => {
      const next = [...prev];
      next[index] = { ...next[index], is_active: !next[index].is_active };
      return next;
    });
  };

  const handleCycleSeatCategory = (index: number) => {
    if (categories.length <= 1) return;
    setCustomSeats((prev) => {
      const next = [...prev];
      const currentCatId = next[index].seat_category_id;
      const currentIndex = categories.findIndex((c) => c.id === currentCatId);
      const nextIndex = (currentIndex + 1) % categories.length;
      next[index] = { ...next[index], seat_category_id: categories[nextIndex].id };
      return next;
    });
  };

  const handleDeployLayout = async () => {
    if (!selectedVenueId) {
      alert("Please select a target venue.");
      return;
    }
    if (customSeats.length === 0) {
      alert("No seats to deploy.");
      return;
    }

    setDeployingLayout(true);
    try {
      const res = await batchConfigureSeats(selectedVenueId, {
        replace: true,
        seats: customSeats,
      });
      setSuccessMessage(
        `🎉 Successfully configured ${res.total_configured} seats (${res.active_capacity} active capacity) for venue!`
      );
      await Promise.all([loadData(), loadVenueSeats(selectedVenueId)]);
    } catch (err: any) {
      alert(`Failed to deploy seat layout: ${err.message}`);
    } finally {
      setDeployingLayout(false);
    }
  };

  const handleClearVenueSeats = async () => {
    if (!selectedVenueId) return;
    if (!window.confirm("Are you sure you want to clear all existing seats for this venue?")) return;

    try {
      await deleteVenueSeats(selectedVenueId);
      setSuccessMessage("Venue seats cleared.");
      await Promise.all([loadData(), loadVenueSeats(selectedVenueId)]);
    } catch (err: any) {
      alert(`Failed to clear seats: ${err.message}`);
    }
  };

  if (authLoading) {
    return (
      <div className="py-24 text-center text-slate-400">
        <RefreshCw className="w-8 h-8 mx-auto animate-spin text-indigo-400 mb-3" />
        <p className="font-semibold text-slate-300">Checking Admin permissions...</p>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center space-x-2 text-xs font-mono font-bold text-amber-400 uppercase tracking-wider">
            <ShieldCheck className="w-4 h-4 text-amber-400" />
            <span>PLATFORM ADMINISTRATION</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight mt-1">
            Venues & Seat Layout Management
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Configure stadiums, theatres, cinema halls, seat categories, and visual seat grids.
          </p>
        </div>

        {/* Role Helper Banner */}
        {user?.role !== "ADMIN" && (
          <div className="bg-amber-500/10 border border-amber-500/30 px-4 py-2.5 rounded-2xl flex items-center gap-3">
            <Info className="w-4 h-4 text-amber-400 shrink-0" />
            <div className="text-xs text-amber-200">
              <span className="font-bold">Logged in as {user?.role || "GUEST"}</span>. Switch to Admin mode to manage live venues.
            </div>
            <button
              onClick={() => switchRole("ADMIN")}
              className="px-3 py-1 bg-amber-500 hover:bg-amber-400 text-black font-bold text-xs rounded-xl transition"
            >
              Switch to ADMIN
            </button>
          </div>
        )}
      </div>

      {/* Success Notification */}
      {successMessage && (
        <div className="p-3.5 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs flex items-center justify-between animate-in fade-in">
          <div className="flex items-center space-x-2">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
            <span className="font-semibold">{successMessage}</span>
          </div>
          <button onClick={() => setSuccessMessage(null)} className="text-emerald-400/70 hover:text-emerald-300 font-bold">
            ✕
          </button>
        </div>
      )}

      {/* Tabs Navigation */}
      <div className="flex border-b border-slate-800 space-x-2 text-xs font-semibold">
        <button
          onClick={() => setActiveTab("venues")}
          className={`pb-3 px-4 flex items-center space-x-2 border-b-2 transition ${
            activeTab === "venues"
              ? "border-amber-400 text-amber-300 font-bold"
              : "border-transparent text-slate-400 hover:text-slate-200"
          }`}
        >
          <Building2 className="w-4 h-4" />
          <span>Venues & Auditoriums ({venues.length})</span>
        </button>

        <button
          onClick={() => setActiveTab("categories")}
          className={`pb-3 px-4 flex items-center space-x-2 border-b-2 transition ${
            activeTab === "categories"
              ? "border-amber-400 text-amber-300 font-bold"
              : "border-transparent text-slate-400 hover:text-slate-200"
          }`}
        >
          <Palette className="w-4 h-4" />
          <span>Seat Categories ({categories.length})</span>
        </button>

        <button
          onClick={() => setActiveTab("layout")}
          className={`pb-3 px-4 flex items-center space-x-2 border-b-2 transition ${
            activeTab === "layout"
              ? "border-amber-400 text-amber-300 font-bold"
              : "border-transparent text-slate-400 hover:text-slate-200"
          }`}
        >
          <Grid3X3 className="w-4 h-4" />
          <span>Interactive Visual Layout Builder</span>
        </button>
      </div>

      {/* =========================================================================
          TAB 1: VENUES MANAGEMENT
          ========================================================================= */}
      {activeTab === "venues" && (
        <div className="space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <h3 className="text-lg font-bold text-white flex items-center">
              <Building2 className="w-5 h-5 mr-2 text-indigo-400" />
              Registered Venues
            </h3>
            <button
              onClick={() => handleOpenVenueModal()}
              className="px-4 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-md shadow-indigo-600/20 transition flex items-center self-start sm:self-auto"
            >
              <Plus className="w-4 h-4 mr-1.5" />
              Add New Venue
            </button>
          </div>

          {loading ? (
            <div className="py-16 text-center text-slate-400 text-xs">
              <RefreshCw className="w-6 h-6 animate-spin mx-auto text-indigo-400 mb-2" />
              Loading venues...
            </div>
          ) : venues.length === 0 ? (
            <div className="py-16 text-center glass-panel rounded-3xl border border-slate-800 p-8 space-y-3">
              <Building2 className="w-10 h-10 mx-auto text-slate-600" />
              <h4 className="text-base font-bold text-slate-300">No venues created yet</h4>
              <p className="text-xs text-slate-500 max-w-sm mx-auto">
                Create a venue to start assigning visual seat layouts and scheduling live concerts or movie screenings.
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {venues.map((venue) => (
                <div
                  key={venue.id}
                  className="glass-panel rounded-3xl p-6 border border-slate-800 bg-[#131A26] shadow-xl hover:border-slate-700 transition flex flex-col justify-between space-y-4"
                >
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                        VENUE ID: {venue.id.slice(0, 8)}...
                      </span>
                      <span className="text-xs font-mono font-bold text-emerald-400 flex items-center">
                        <Users className="w-3.5 h-3.5 mr-1" />
                        {venue.total_capacity} seats
                      </span>
                    </div>

                    <h4 className="text-lg font-bold text-white tracking-tight">{venue.name}</h4>

                    <div className="text-xs text-slate-400 space-y-1">
                      <p className="flex items-center">
                        <MapPin className="w-3.5 h-3.5 mr-1.5 text-rose-400 shrink-0" />
                        <span>{venue.address}, {venue.city}</span>
                      </p>
                      {venue.state && (
                        <p className="text-slate-500 pl-5">
                          {venue.state}, {venue.country}
                        </p>
                      )}
                    </div>
                  </div>

                  <div className="pt-3 border-t border-slate-800/80 flex items-center justify-between gap-2">
                    <button
                      onClick={() => {
                        setSelectedVenueId(venue.id);
                        setActiveTab("layout");
                      }}
                      className="px-3 py-1.5 rounded-xl text-xs font-semibold text-indigo-300 hover:text-white bg-indigo-950/60 border border-indigo-800/50 hover:bg-indigo-900 transition flex items-center"
                    >
                      <Grid3X3 className="w-3.5 h-3.5 mr-1.5" />
                      Configure Layout
                    </button>

                    <div className="flex items-center space-x-1">
                      <button
                        onClick={() => handleOpenVenueModal(venue)}
                        className="p-2 text-slate-400 hover:text-slate-200 rounded-lg hover:bg-slate-800 transition"
                        title="Edit Venue Details"
                      >
                        <Edit2 className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={() => handleDeleteVenue(venue)}
                        className="p-2 text-rose-400 hover:text-rose-300 rounded-lg hover:bg-rose-950/30 transition"
                        title="Delete Venue"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* =========================================================================
          TAB 2: SEAT CATEGORIES MANAGEMENT
          ========================================================================= */}
      {activeTab === "categories" && (
        <div className="space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h3 className="text-lg font-bold text-white flex items-center">
                <Palette className="w-5 h-5 mr-2 text-violet-400" />
                Seat Categories & Tiers
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Categories are assigned to seats in venues and priced per-event by organisers (e.g. VIP, Premium, Standard).
              </p>
            </div>
            <button
              onClick={() => setShowCategoryModal(true)}
              className="px-4 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-violet-600 to-purple-600 hover:from-violet-500 hover:to-purple-500 shadow-md shadow-violet-600/20 transition flex items-center self-start sm:self-auto"
            >
              <Plus className="w-4 h-4 mr-1.5" />
              Add New Category
            </button>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {categories.map((cat) => (
              <div
                key={cat.id}
                className="glass-panel rounded-3xl p-6 border border-slate-800 bg-[#131A26] shadow-xl space-y-4"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2.5">
                    <span
                      className="w-4 h-4 rounded-full shadow-md"
                      style={{ backgroundColor: cat.color_code }}
                    />
                    <h4 className="text-base font-bold text-white tracking-tight">{cat.name}</h4>
                  </div>
                  <span className="text-[11px] font-mono font-bold text-slate-400 bg-slate-900 px-2.5 py-0.5 rounded-lg border border-slate-800">
                    {cat.color_code}
                  </span>
                </div>

                <p className="text-xs text-slate-400 min-h-[36px]">
                  {cat.description || "No description provided."}
                </p>

                <div className="pt-3 border-t border-slate-800 text-[11px] font-mono text-slate-500 flex justify-between">
                  <span>ID: {cat.id.slice(0, 8)}...</span>
                  <span>Created: {new Date(cat.created_at).toLocaleDateString()}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* =========================================================================
          TAB 3: INTERACTIVE VISUAL SEAT LAYOUT BUILDER
          ========================================================================= */}
      {activeTab === "layout" && (
        <div className="space-y-6">
          {/* Target Venue & Dimension Controls */}
          <div className="glass-panel rounded-3xl p-6 border border-slate-800 bg-[#131A26] space-y-6">
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-slate-800 pb-4">
              <div>
                <h3 className="text-lg font-bold text-white flex items-center">
                  <Grid3X3 className="w-5 h-5 mr-2 text-indigo-400" />
                  Visual Seat Grid Configurator
                </h3>
                <p className="text-xs text-slate-400 mt-0.5">
                  Design the 2D visual seating map, assign categories per row, and batch deploy to the database.
                </p>
              </div>

              {/* Venue Selector */}
              <div className="flex items-center space-x-3">
                <span className="text-xs font-semibold text-slate-400 uppercase font-mono">
                  Target Venue:
                </span>
                <select
                  value={selectedVenueId}
                  onChange={(e) => setSelectedVenueId(e.target.value)}
                  className="bg-[#0B0F17] border border-slate-700/80 rounded-xl px-3 py-2 text-xs font-bold text-slate-100 focus:outline-none focus:border-indigo-500 transition cursor-pointer"
                >
                  {venues.map((v) => (
                    <option key={v.id} value={v.id}>
                      {v.name} ({v.city}) - currently {v.total_capacity} seats
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {/* Grid Matrix Controls */}
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 text-xs font-mono">
              <div className="bg-[#0B0F17] p-3.5 rounded-2xl border border-slate-800 space-y-2">
                <label className="text-slate-400 font-semibold block uppercase text-[11px]">
                  Rows (Depth: A to {rowLetters[gridRowsCount - 1]})
                </label>
                <input
                  type="number"
                  min="1"
                  max="16"
                  value={gridRowsCount}
                  onChange={(e) => setGridRowsCount(Math.max(1, Math.min(16, parseInt(e.target.value) || 1)))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-white font-bold"
                />
              </div>

              <div className="bg-[#0B0F17] p-3.5 rounded-2xl border border-slate-800 space-y-2">
                <label className="text-slate-400 font-semibold block uppercase text-[11px]">
                  Columns (Width per row)
                </label>
                <input
                  type="number"
                  min="1"
                  max="24"
                  value={gridColsCount}
                  onChange={(e) => setGridColsCount(Math.max(1, Math.min(24, parseInt(e.target.value) || 1)))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-white font-bold"
                />
              </div>

              <div className="bg-[#0B0F17] p-3.5 rounded-2xl border border-slate-800 space-y-2">
                <span className="text-slate-400 font-semibold block uppercase text-[11px]">
                  Configured Layout Stats
                </span>
                <div className="font-bold text-white text-sm">
                  {customSeats.length} seats ({customSeats.filter((s) => s.is_active).length} active)
                </div>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={handleDeployLayout}
                  disabled={deployingLayout || customSeats.length === 0}
                  className="flex-1 py-3.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 shadow-lg shadow-emerald-600/30 transition flex items-center justify-center disabled:opacity-40"
                >
                  {deployingLayout ? (
                    <RefreshCw className="w-4 h-4 animate-spin" />
                  ) : (
                    <>
                      <Sparkles className="w-4 h-4 mr-1.5" />
                      Save & Deploy Layout
                    </>
                  )}
                </button>

                {venueSeats.length > 0 && (
                  <button
                    onClick={handleClearVenueSeats}
                    className="p-3.5 rounded-xl text-rose-400 hover:text-rose-200 bg-rose-950/30 border border-rose-800/40 hover:bg-rose-900/40 transition"
                    title="Clear existing venue seats"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </div>
            </div>

            {/* Row Category Assignment Selector */}
            <div className="space-y-3 pt-2">
              <span className="text-xs font-semibold text-slate-300 uppercase font-mono block">
                Assign Category Per Row:
              </span>
              <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 gap-2">
                {rowLetters.slice(0, gridRowsCount).map((row) => (
                  <div key={row} className="bg-[#0B0F17] p-2 rounded-xl border border-slate-800 text-xs">
                    <span className="font-bold font-mono text-slate-200 block mb-1">Row {row}</span>
                    <select
                      value={rowCategoryMapping[row] || categories[0]?.id || ""}
                      onChange={(e) => {
                        const newCatId = e.target.value;
                        setRowCategoryMapping((prev) => ({ ...prev, [row]: newCatId }));
                      }}
                      className="w-full bg-slate-900 border border-slate-700 rounded px-1.5 py-1 text-[11px] text-slate-200"
                    >
                      {categories.map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.name}
                        </option>
                      ))}
                    </select>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Interactive Seat Map Canvas Preview */}
          <div className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] space-y-6">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <h4 className="text-base font-bold text-white flex items-center">
                  <Eye className="w-4 h-4 mr-2 text-indigo-400" />
                  Interactive Layout Canvas Preview
                </h4>
                <p className="text-xs text-slate-400 mt-0.5">
                  Click a seat to cycle its category or toggle active status.
                </p>
              </div>

              {/* Category Legend */}
              <div className="flex flex-wrap items-center gap-3 text-xs">
                {categories.map((cat) => (
                  <div key={cat.id} className="flex items-center space-x-1.5">
                    <span className="w-3 h-3 rounded" style={{ backgroundColor: cat.color_code }} />
                    <span className="text-slate-300 font-medium">{cat.name}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Screen / Stage Direction Indicator */}
            <div className="py-2 text-center space-y-2">
              <div className="h-2 w-2/3 mx-auto rounded-full bg-gradient-to-r from-transparent via-indigo-500 to-transparent shadow-[0_0_20px_rgba(99,102,241,0.8)]" />
              <span className="text-[10px] font-mono uppercase tracking-widest text-slate-500">
                SCREEN / STAGE FRONT
              </span>
            </div>

            {/* 2D Visual Seat Grid */}
            <div className="overflow-x-auto py-6 flex flex-col items-center gap-2">
              {rowLetters.slice(0, gridRowsCount).map((rowLabel) => {
                const rowSeats = customSeats.filter((s) => s.row_label === rowLabel);

                return (
                  <div key={rowLabel} className="flex items-center gap-2">
                    <span className="w-6 text-center text-xs font-mono font-bold text-slate-500">
                      {rowLabel}
                    </span>

                    <div className="flex items-center gap-1.5">
                      {rowSeats.map((seat) => {
                        const globalIndex = customSeats.findIndex(
                          (s) => s.row_label === seat.row_label && s.seat_number === seat.seat_number
                        );
                        const cat = categories.find((c) => c.id === seat.seat_category_id);
                        const catColor = cat ? cat.color_code : "#3B82F6";

                        return (
                          <button
                            key={`${seat.row_label}-${seat.seat_number}`}
                            onClick={() => handleCycleSeatCategory(globalIndex)}
                            onContextMenu={(e) => {
                              e.preventDefault();
                              handleToggleSeatActive(globalIndex);
                            }}
                            title={`Row ${seat.row_label}, Seat ${seat.seat_number} - ${cat?.name || "Category"} (Right-click to toggle inactive)`}
                            className={`w-7 h-7 sm:w-8 sm:h-8 rounded-lg text-[10px] font-bold font-mono transition-all flex items-center justify-center text-white shadow-sm ${
                              seat.is_active ? "hover:scale-105" : "opacity-20 line-through bg-slate-800"
                            }`}
                            style={{ backgroundColor: seat.is_active ? catColor : undefined }}
                          >
                            {seat.seat_number}
                          </button>
                        );
                      })}
                    </div>

                    <span className="w-6 text-center text-xs font-mono font-bold text-slate-500">
                      {rowLabel}
                    </span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* =========================================================================
          VENUE MODAL (CREATE / EDIT)
          ========================================================================= */}
      {showVenueModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-lg bg-[#131A26] border border-slate-700 rounded-3xl p-6 sm:p-8 space-y-6 shadow-2xl animate-in zoom-in-95">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <h3 className="text-lg font-bold text-white flex items-center">
                <Building2 className="w-5 h-5 mr-2 text-indigo-400" />
                {editingVenue ? "Edit Venue" : "Create New Venue"}
              </h3>
              <button onClick={() => setShowVenueModal(false)} className="text-slate-400 hover:text-white font-bold">
                ✕
              </button>
            </div>

            <form onSubmit={handleSaveVenue} className="space-y-4 text-xs font-mono">
              <div className="space-y-1">
                <label className="text-slate-300 font-semibold block">Venue Name *</label>
                <input
                  type="text"
                  required
                  value={venueName}
                  onChange={(e) => setVenueName(e.target.value)}
                  placeholder="e.g. Grand Arena, Dolby Cinema Hall 1"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#0B0F17] border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="space-y-1">
                <label className="text-slate-300 font-semibold block">Street Address *</label>
                <input
                  type="text"
                  required
                  value={venueAddress}
                  onChange={(e) => setVenueAddress(e.target.value)}
                  placeholder="e.g. 100 Arena Blvd, Downtown"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#0B0F17] border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-slate-300 font-semibold block">City *</label>
                  <input
                    type="text"
                    required
                    value={venueCity}
                    onChange={(e) => setVenueCity(e.target.value)}
                    placeholder="e.g. San Francisco"
                    className="w-full px-3.5 py-2.5 rounded-xl bg-[#0B0F17] border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-slate-300 font-semibold block">State / Province</label>
                  <input
                    type="text"
                    value={venueState}
                    onChange={(e) => setVenueState(e.target.value)}
                    placeholder="e.g. California"
                    className="w-full px-3.5 py-2.5 rounded-xl bg-[#0B0F17] border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="flex gap-3 pt-4 border-t border-slate-800">
                <button
                  type="submit"
                  disabled={savingVenue}
                  className="flex-1 py-3 rounded-xl font-bold text-xs text-white bg-indigo-600 hover:bg-indigo-500 shadow-md shadow-indigo-600/30 transition flex items-center justify-center"
                >
                  {savingVenue ? <RefreshCw className="w-4 h-4 animate-spin" /> : "Save Venue"}
                </button>
                <button
                  type="button"
                  onClick={() => setShowVenueModal(false)}
                  className="px-5 py-3 rounded-xl font-semibold text-xs text-slate-300 hover:text-white bg-slate-800 border border-slate-700 transition"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* =========================================================================
          CATEGORY MODAL (CREATE)
          ========================================================================= */}
      {showCategoryModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-md bg-[#131A26] border border-slate-700 rounded-3xl p-6 sm:p-8 space-y-6 shadow-2xl animate-in zoom-in-95">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <h3 className="text-lg font-bold text-white flex items-center">
                <Palette className="w-5 h-5 mr-2 text-violet-400" />
                Create Seat Category
              </h3>
              <button onClick={() => setShowCategoryModal(false)} className="text-slate-400 hover:text-white font-bold">
                ✕
              </button>
            </div>

            <form onSubmit={handleSaveCategory} className="space-y-4 text-xs font-mono">
              <div className="space-y-1">
                <label className="text-slate-300 font-semibold block">Category Name *</label>
                <input
                  type="text"
                  required
                  value={categoryName}
                  onChange={(e) => setCategoryName(e.target.value)}
                  placeholder="e.g. VIP Front Row, Club Lounge, Standard"
                  className="w-full px-3.5 py-2.5 rounded-xl bg-[#0B0F17] border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-violet-500"
                />
              </div>

              <div className="space-y-1">
                <label className="text-slate-300 font-semibold block">Color Code (Hex) *</label>
                <div className="flex items-center space-x-3">
                  <input
                    type="color"
                    value={categoryColor}
                    onChange={(e) => setCategoryColor(e.target.value)}
                    className="w-10 h-10 rounded-lg cursor-pointer bg-transparent border-0"
                  />
                  <input
                    type="text"
                    required
                    value={categoryColor}
                    onChange={(e) => setCategoryColor(e.target.value)}
                    placeholder="#3B82F6"
                    className="flex-1 px-3.5 py-2.5 rounded-xl bg-[#0B0F17] border border-slate-700 text-white font-mono uppercase focus:outline-none focus:border-violet-500"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-slate-300 font-semibold block">Description</label>
                <textarea
                  rows={2}
                  value={categoryDesc}
                  onChange={(e) => setCategoryDesc(e.target.value)}
                  placeholder="Optional description of seating amenities or perks..."
                  className="w-full px-3.5 py-2 rounded-xl bg-[#0B0F17] border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-violet-500"
                />
              </div>

              <div className="flex gap-3 pt-4 border-t border-slate-800">
                <button
                  type="submit"
                  disabled={savingCategory}
                  className="flex-1 py-3 rounded-xl font-bold text-xs text-white bg-violet-600 hover:bg-violet-500 shadow-md shadow-violet-600/30 transition flex items-center justify-center"
                >
                  {savingCategory ? <RefreshCw className="w-4 h-4 animate-spin" /> : "Create Category"}
                </button>
                <button
                  type="button"
                  onClick={() => setShowCategoryModal(false)}
                  className="px-5 py-3 rounded-xl font-semibold text-xs text-slate-300 hover:text-white bg-slate-800 border border-slate-700 transition"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
