#include "feedbackservice.h"

#include <QtCore/QUrl>
#include <QtGui/QDesktopServices>

FeedbackService::FeedbackService(QObject *parent)
    : QObject(parent)
{ }

void FeedbackService::openUrl(const QString &url)
{
    QDesktopServices::openUrl(QUrl(url));
}
