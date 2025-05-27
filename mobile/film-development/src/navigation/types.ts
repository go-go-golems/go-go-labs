import { NativeStackScreenProps } from '@react-navigation/native-stack';
import { Roll, ProcessStep } from '../types';

export type RootStackParamList = {
  Home: undefined;
  Timer: {
    rolls: Roll[];
    processSteps: ProcessStep[];
  };
};

export type HomeScreenProps = NativeStackScreenProps<RootStackParamList, 'Home'>;
export type TimerScreenProps = NativeStackScreenProps<RootStackParamList, 'Timer'>;