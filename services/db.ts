
import { UserProfile, QuizResult } from "../types";

const KEYS = {
  USER: 'edupulse_user',
  PROFILE: 'edupulse_profile',
  HISTORY: 'edupulse_quiz_history'
};

export const db = {
  saveUser: (email: string) => {
    localStorage.setItem(KEYS.USER, JSON.stringify({ email, isLoggedIn: true }));
  },
  
  getUser: () => {
    const data = localStorage.getItem(KEYS.USER);
    return data ? JSON.parse(data) : null;
  },

  logout: () => {
    localStorage.removeItem(KEYS.USER);
  },

  saveProfile: (profile: UserProfile) => {
    localStorage.setItem(KEYS.PROFILE, JSON.stringify(profile));
  },

  getProfile: (): UserProfile | null => {
    const data = localStorage.getItem(KEYS.PROFILE);
    return data ? JSON.parse(data) : null;
  },

  saveQuizResult: (result: QuizResult) => {
    const history = db.getQuizHistory();
    history.push({ ...result, timestamp: new Date().toISOString() });
    localStorage.setItem(KEYS.HISTORY, JSON.stringify(history));
  },

  getQuizHistory: (): any[] => {
    const data = localStorage.getItem(KEYS.HISTORY);
    return data ? JSON.parse(data) : [];
  }
};
