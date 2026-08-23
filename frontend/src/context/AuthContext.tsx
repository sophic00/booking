"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { User, UserRole } from "../lib/types";
import { authLogin, authRegister } from "../lib/api";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, pass: string) => Promise<void>;
  register: (data: { email: string; password: string; full_name: string; phone?: string; role?: string }) => Promise<void>;
  logout: () => void;
  switchRole: (role: UserRole) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    try {
      const storedToken = localStorage.getItem("velvet_auth_token");
      const storedUser = localStorage.getItem("velvet_user_profile");
      if (storedToken && storedUser) {
        setToken(storedToken);
        setUser(JSON.parse(storedUser));
      } else {
        // Provide a default active customer session for seamless previewing
        const demoUser: User = {
          id: "usr-demo-01",
          email: "alex.rivera@example.com",
          full_name: "Alex Rivera",
          phone: "+1 (555) 234-8901",
          role: "CUSTOMER",
          created_at: new Date().toISOString(),
        };
        setUser(demoUser);
        setToken("demo-token-123");
        localStorage.setItem("velvet_user_profile", JSON.stringify(demoUser));
        localStorage.setItem("velvet_auth_token", "demo-token-123");
      }
    } catch {
      // ignore
    } finally {
      setIsLoading(false);
    }
  }, []);

  const login = async (email: string, pass: string) => {
    setIsLoading(true);
    try {
      const res = await authLogin(email, pass);
      setUser(res.user);
      setToken(res.token);
      localStorage.setItem("velvet_auth_token", res.token);
      localStorage.setItem("velvet_user_profile", JSON.stringify(res.user));
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (data: { email: string; password: string; full_name: string; phone?: string; role?: string }) => {
    setIsLoading(true);
    try {
      const res = await authRegister(data);
      setUser(res.user);
      setToken(res.token);
      localStorage.setItem("velvet_auth_token", res.token);
      localStorage.setItem("velvet_user_profile", JSON.stringify(res.user));
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem("velvet_auth_token");
    localStorage.removeItem("velvet_user_profile");
  };

  const switchRole = (newRole: UserRole) => {
    if (!user) return;
    const updated = { ...user, role: newRole };
    setUser(updated);
    localStorage.setItem("velvet_user_profile", JSON.stringify(updated));
  };

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, register, logout, switchRole }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
