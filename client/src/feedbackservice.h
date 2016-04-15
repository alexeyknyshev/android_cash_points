#ifndef FEEDBACKSERVICE_H
#define FEEDBACKSERVICE_H

#include <QtCore/QObject>

class FeedbackService : public QObject
{
    Q_OBJECT

public:
    explicit FeedbackService(QObject *parent = 0);

    Q_INVOKABLE void openUrl(const QString &url);
};

#endif // FEEDBACKSERVICE_H
