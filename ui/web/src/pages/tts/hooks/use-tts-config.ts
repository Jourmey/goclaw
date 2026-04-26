import { useConfig } from "@/pages/config/hooks/use-config";
import { useHttp } from "@/hooks/use-ws";

export interface SynthesizeParams {
  text: string;
  provider: string;
  voice_id?: string;
  model_id?: string;
  params?: Record<string, unknown>;
}

export interface TtsConfig {
  provider?: string;
  auto?: string;
}

/**
 * Hook for accessing TTS configuration and synthesis functionality.
 * Note: TTS feature was removed in v3.x but UI components remain for future use.
 */
export function useTtsConfig() {
  const { config } = useConfig();
  const http = useHttp();

  const tts: TtsConfig = (config?.tts as TtsConfig) ?? {};

  const synthesize = async (params: SynthesizeParams): Promise<Blob> => {
    // TTS endpoint not yet implemented in v3.x
    // This is a placeholder for future TTS API integration
    throw new Error("TTS synthesis is not available in this version");
  };

  return {
    tts,
    synthesize,
  };
}
