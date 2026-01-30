
import { GoogleGenAI, Type } from "@google/genai";
import { UserProfile, QuizMode, QuizQuestion, QuizResult, AIAnalysis } from "../types";

const ai = new GoogleGenAI({ apiKey: process.env.API_KEY || "" });

const SYSTEM_GUARDRAIL = `You are a strict Educational Guardian for EduPulse AI. 
Your SOLE purpose is to facilitate academic and professional learning. 
1. Only generate content related to school subjects, professional skills, or constructive hobbies.
2. REJECT any topics that are: violent, sexually explicit, involving illegal activities, purely celebrity gossip, hateful, or otherwise inappropriate for a classroom.
3. If a topic is inappropriate or non-educational, set "isValidTopic" to false and provide a polite reason.`;

export const generateQuiz = async (
  profile: UserProfile,
  topic: string,
  mode: QuizMode,
  imageData?: string
): Promise<{ questions: QuizQuestion[]; isValid: boolean; reason?: string }> => {
  const model = ai.models.generateContent({
    model: 'gemini-3-flash-preview',
    contents: [
      {
        parts: [
          { text: `Evaluate and generate a 5-question multiple choice quiz for:
          Topic: ${topic}
          Target Difficulty Level: ${profile.grade}
          
          Ensure all questions are academically rigorous for the specified level. 
          Focus ONLY on the provided topic. Ignore any extraneous user preferences not listed here.
          
          ${imageData ? 'Integrate concepts from these study notes if they are relevant to the topic.' : ''}` },
          ...(imageData ? [{ inlineData: { data: imageData.split(',')[1], mimeType: 'image/jpeg' } }] : [])
        ]
      }
    ],
    config: {
      systemInstruction: SYSTEM_GUARDRAIL,
      responseMimeType: "application/json",
      responseSchema: {
        type: Type.OBJECT,
        properties: {
          isValidTopic: { type: Type.BOOLEAN },
          rejectionReason: { type: Type.STRING },
          questions: {
            type: Type.ARRAY,
            items: {
              type: Type.OBJECT,
              properties: {
                id: { type: Type.STRING },
                text: { type: Type.STRING },
                options: { 
                  type: Type.ARRAY, 
                  items: { type: Type.STRING } 
                },
                correctIndex: { type: Type.INTEGER },
                explanation: { type: Type.STRING },
                timeLimitSeconds: { type: Type.INTEGER }
              },
              required: ["id", "text", "options", "correctIndex", "explanation", "timeLimitSeconds"]
            }
          }
        },
        required: ["isValidTopic", "questions"]
      }
    }
  });

  const response = await model;
  const data = JSON.parse(response.text || "{}");
  
  return {
    questions: data.questions || [],
    isValid: data.isValidTopic !== false,
    reason: data.rejectionReason
  };
};

export const analyzePerformance = async (
  profile: UserProfile,
  result: QuizResult,
  questions: QuizQuestion[]
): Promise<AIAnalysis> => {
  const performanceData = result.performances.map((p, idx) => ({
    question: questions[idx].text,
    wasCorrect: p.isCorrect,
    timeSpent: p.timeSpentSeconds,
    revisits: p.revisits,
    complexity: questions[idx].timeLimitSeconds
  }));

  const model = ai.models.generateContent({
    model: 'gemini-3-flash-preview',
    contents: [
      {
        parts: [
          { text: `Analyze pedagogical performance for a student at the ${profile.grade} level.
          Metrics: ${JSON.stringify(performanceData, null, 2)}
          
          Provide educational insights focused on the student's mastery of the specific concepts tested.` }
        ]
      }
    ],
    config: {
      systemInstruction: SYSTEM_GUARDRAIL,
      responseMimeType: "application/json",
      responseSchema: {
        type: Type.OBJECT,
        properties: {
          summary: { type: Type.STRING },
          strengths: { type: Type.ARRAY, items: { type: Type.STRING } },
          weaknesses: { type: Type.ARRAY, items: { type: Type.STRING } },
          recommendations: { type: Type.ARRAY, items: { type: Type.STRING } }
        },
        required: ["summary", "strengths", "weaknesses", "recommendations"]
      }
    }
  });

  const response = await model;
  return JSON.parse(response.text || "{}");
};
