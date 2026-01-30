
export enum QuizMode {
  TIMED_QUESTION = 'TIMED_QUESTION',
  OVERALL_TIMED = 'OVERALL_TIMED'
}

export interface UserProfile {
  name: string;
  grade: string;
}

export interface QuizQuestion {
  id: string;
  text: string;
  options: string[];
  correctIndex: number;
  explanation: string;
  timeLimitSeconds: number;
}

export interface QuestionPerformance {
  questionId: string;
  timeSpentSeconds: number;
  attempts: number;
  isCorrect: boolean;
  revisits: number;
}

export interface QuizResult {
  totalQuestions: number;
  correctAnswers: number;
  totalTimeSpent: number;
  performances: QuestionPerformance[];
  mode: QuizMode;
}

export interface AIAnalysis {
  summary: string;
  strengths: string[];
  weaknesses: string[];
  recommendations: string[];
}
