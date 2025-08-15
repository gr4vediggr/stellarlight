import { create } from "zustand";

interface UserState {
  username: string | null;
  setUsername: (u: string | null) => void;
}

export const useUserStore = create<UserState>((set) => ({
  username: null,
  setUsername: (u) => set({ username: u })
}));
