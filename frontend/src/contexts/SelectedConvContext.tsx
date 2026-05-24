"use client";

import React, { createContext, useContext, useState, type ReactNode } from "react";
import type { Conversation } from "@/types";

interface SelectedConvContextType {
  selectedConv: Conversation | null;
  setSelectedConv: (conv: Conversation | null) => void;
}

const SelectedConvContext = createContext<SelectedConvContextType>({
  selectedConv: null,
  setSelectedConv: () => {},
});

export function SelectedConvProvider({ children }: { children: ReactNode }) {
  const [selectedConv, setSelectedConv] = useState<Conversation | null>(null);
  return (
    <SelectedConvContext.Provider value={{ selectedConv, setSelectedConv }}>
      {children}
    </SelectedConvContext.Provider>
  );
}

export function useSelectedConv() {
  return useContext(SelectedConvContext);
}
