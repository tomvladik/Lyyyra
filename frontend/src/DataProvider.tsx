import { PropsWithChildren, createContext, useContext, useEffect, useState } from "react";
import { AppStatus, isEqualAppStatus } from "./AppStatus";

interface IContextValue {
    data: AppStatus;
    updateData: (newData: AppStatus) => void;
}

export const DataContext = createContext<IContextValue>(null!);

export default function useDataContext() {
    const context = useContext(DataContext);
    if (!context) {
        throw new Error(
            "useDataContext must be used within the DataContextProvider"
        );
    }
    return context;
}