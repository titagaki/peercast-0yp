/** Asia/Tokyo の今日の日付を YYYYMMDD 形式で返す */
export function todayYYYYMMDD(): string {
  // sv-SE locale produces YYYY-MM-DD; use Asia/Tokyo to avoid UTC date shift
  return new Date().toLocaleDateString('sv-SE', { timeZone: 'Asia/Tokyo' }).replace(/-/g, '')
}
