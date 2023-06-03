/* eslint-disable react/no-array-index-key */

import { useCallback } from 'react'

import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'

import { ChatHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

import { View } from './NavBar'
import { VSCodeWrapper } from './utils/VSCodeApi'

import chatStyles from './Chat.module.css'
import styles from './UserHistory.module.css'

interface HistoryProps {
    userHistory: ChatHistory | null
    setUserHistory: (history: ChatHistory | null) => void
    setInputHistory: (inputHistory: string[] | []) => void
    setView: (view: View | undefined) => void
    vscodeAPI: VSCodeWrapper
}

export const UserHistory: React.FunctionComponent<React.PropsWithChildren<HistoryProps>> = ({
    userHistory,
    setUserHistory,
    setInputHistory,
    setView,
    vscodeAPI,
}) => {
    const onDeleteHistoryClick = useCallback(
        (chatID: string): void => {
            if (userHistory) {
                delete userHistory[chatID]
                setUserHistory({ ...userHistory })
                vscodeAPI.postMessage({ command: 'deleteHistory', chatID })
            }
        },
        [userHistory, setUserHistory, vscodeAPI]
    )

    const onRemoveHistoryClick = useCallback(() => {
        if (userHistory) {
            vscodeAPI.postMessage({ command: 'removeHistory' })
            setUserHistory(null)
            setInputHistory([])
        }
    }, [setInputHistory, userHistory, setUserHistory, vscodeAPI])

    function restoreMetadata(chatID: string): void {
        vscodeAPI.postMessage({ command: 'restoreHistory', chatID })
        setView('chat')
    }

    // Fix this funcction

    const findTimeDifference = (interactionTime: string): string => {
        const date = new Date(interactionTime)
        const now = new Date()
        const diff = now.getTime() - date.getTime()

        const hours = Math.floor(diff / (60 * 60 * 1000))
        const minutes = Math.floor(diff / (60 * 1000)) % 60
        const seconds = Math.floor(diff / 1000) % 60

        return `${hours}h ${minutes}m ${seconds}s`
    }

    return (
        <div className={chatStyles.innerContainer}>
            <div className={chatStyles.nonTranscriptContainer}>
                <div className={styles.headingContainer}>
                    <h3>Chat History</h3>
                    <div className={styles.clearButtonContainer}>
                        <VSCodeButton
                            className={styles.clearButton}
                            type="button"
                            onClick={onRemoveHistoryClick}
                            disabled={userHistory === null}
                        >
                            Clear History
                        </VSCodeButton>
                    </div>
                </div>
                <div className={styles.itemsContainer}>
                    {userHistory &&
                        [...Object.entries(userHistory)]
                            .sort(
                                (a, b) =>
                                    +new Date(b[1].lastInteractionTimestamp) - +new Date(a[1].lastInteractionTimestamp)
                            )
                            .map(chat => {
                                const lastMessage = chat[1].interactions[chat[1].interactions.length - 1].humanMessage
                                if (!lastMessage?.displayText) {
                                    return null
                                }

                                return (
                                    <VSCodeButton
                                        key={chat[0]}
                                        className={styles.itemButton}
                                        onClick={() => restoreMetadata(chat[0])}
                                        type="button"
                                    >
                                        <div className={styles.itemButtonInnerContainer}>
                                            <div className={styles.historyItem}>
                                                <div className={styles.itemDate}>
                                                    {findTimeDifference(new Date(chat[0]).toLocaleString())}
                                                </div>
                                                <div className={styles.itemDelete}>
                                                    <VSCodeButton
                                                        appearance="icon"
                                                        type="button"
                                                        onClick={event => {
                                                            onDeleteHistoryClick(chat[0])
                                                            event.stopPropagation()
                                                        }}
                                                    >
                                                        <i className="codicon codicon-trash" />
                                                    </VSCodeButton>
                                                </div>
                                            </div>
                                            <div className={styles.itemLastMessage}>{lastMessage.displayText}</div>
                                        </div>
                                    </VSCodeButton>
                                )
                            })}
                </div>
            </div>
        </div>
    )
}
